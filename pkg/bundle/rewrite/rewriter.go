package rewrite

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/bundle/replacement"
	"github.com/deis/duffle/pkg/bundle/rewrite/config"
)

// // Rewriter is an interface that is used to rewrite elements of a CNAB bundle
// type Rewriter interface {
// 	Rewrite(contents string, image string, ref bundle.LocationRef) (string, error)
// 	ReplaceRepository(qualifiedImage string, repository string) (string, error)
// 	TagImage(ctx context.Context, oldImage string, newImage string) error
// }

// Rewriter is used to rewrite elements of a CNAB bundle
type Rewriter struct {
	Context *config.Context
	Writer  io.Writer
}

// NewRewriter returns a bundle rewriter
func New(w io.Writer) (*Rewriter, error) {
	context, err := config.New()
	if err != nil {
		return nil, err
	}
	r := &Rewriter{
		Context: context,
		Writer:  w,
	}
	return r, nil
}

//Rewrite updates all image references in a bundle to replace Docker registries
func (r *Rewriter) Rewrite(ctx context.Context, bundle *bundle.Bundle, newRegistry string) error {
	fmt.Fprintf(r.Writer, "Tagging Images\n")
	// for all the referenced images, retag them
	err := r.tagImages(ctx, bundle, newRegistry)
	if err != nil {
		fmt.Fprintf(r.Writer, "error tagging images: %s", err)
		return err
	}
	//for each invocation image, start a container and copy the file system
	for i, invocationImage := range bundle.InvocationImages {
		if err := r.updateInvocationImage(ctx, &invocationImage, bundle.Images, newRegistry); err != nil {
			return err
		}
		bundle.InvocationImages[i] = invocationImage
	}
	//then update the image contents with the refs
	//then commit it
	return nil
}

func (r *Rewriter) tagImages(ctx context.Context, bundle *bundle.Bundle, repo string) error {
	for i, image := range bundle.Images {
		if image.Image == "" {
			return fmt.Errorf("image field not set")
		}
		err := r.pullImageIfNeeded(ctx, image.Image)
		if err != nil {
			return err
		}
		newImage, err := r.replaceRepository(image.Image, repo)
		if err != nil {
			return fmt.Errorf("unable to update image %s: %s", image.Image, err)
		}
		err = r.tagImage(ctx, image.Image, newImage)
		if err != nil {
			return fmt.Errorf("unable to retag image %s: %s", image.Image, err)
		}
		fmt.Fprintf(r.Writer, "Retagged '%s' to '%s'\n", image.Image, newImage)
		bundle.Images[i].Image = newImage
	}
	return nil
}

func (r *Rewriter) updateInvocationImage(ctx context.Context, ii *bundle.InvocationImage, images []bundle.Image, newRegistry string) error {
	// If the bundle's referenced images is empty, we can simply retag the image
	err := r.pullImageIfNeeded(ctx, ii.Image)
	if err != nil {
		return err
	}
	newImage, err := r.replaceRepository(ii.Image, newRegistry)
	if len(images) == 0 {
		if err != nil {
			return fmt.Errorf("unable to update invocation image name %s: %s", ii.Image, err)
		}
		err = r.tagImage(ctx, ii.Image, newImage)
		if err != nil {
			return fmt.Errorf("unable to retag invocation %s: %s", ii.Image, err)
		}
		ii.Image = newImage
		return nil
	}
	// The bundle's referenced images is not empty, so we actually need to update the
	// invocation image
	err = r.updateImageReferences(ctx, ii, images, newRegistry)
	if err != nil {
		return err
	}
	ii.Image = newImage
	return nil
}

func (r *Rewriter) updateImageReferences(ctx context.Context, ii *bundle.InvocationImage, images []bundle.Image, newRepo string) error {
	newImage, err := r.replaceRepository(ii.Image, newRepo)
	if err != nil {
		return err
	}
	tmpDir, err := r.Context.FileSystem.TempDir("", "cnab")
	if err != nil {
		return fmt.Errorf("couldn't get a temp dir")
	}
	containerID, err := r.createContainer(ctx, ii.Image)
	if err != nil {
		return err
	}

	defer func() {
		r.Context.FileSystem.RemoveAll(tmpDir)
		err := r.stopContainer(ctx, containerID)
		if err != nil {
			fmt.Fprintf(r.Writer, "couldn't remove container after rewrite: %s", err)
		}
	}()

	err = r.copyFromContainerToFileSystem(ctx, containerID, tmpDir)
	if err != nil {
		return err
	}
	for _, image := range images {
		for _, ref := range image.Refs {
			contents, fi, err := r.loadReferenceFile(filepath.Join(tmpDir, ref.Path))
			if err != nil {
				return err
			}
			contents, err = r.rewriteFile(contents, newImage, ref)
			if err != nil {
				return err
			}
			err = r.writeReferenceFile([]byte(contents), fi, filepath.Join(tmpDir, ref.Path))
			if err != nil {
				return err
			}
		}
	}
	err = r.copyToContainerFromFileSystem(ctx, containerID, tmpDir)
	if err != nil {
		return err
	}
	id, err := r.commitContainer(ctx, containerID, newImage)
	if err != nil {
		return err
	}
	fmt.Fprintf(r.Writer, "Rewrote invocation image %s to %s\n", ii.Image, newImage)
	fmt.Fprintf(r.Writer, "New image id : %s\n", id)
	ii.Image = newImage
	ii.Digest = id
	return nil
}

func (r *Rewriter) rewriteFile(contents string, image string, ref bundle.LocationRef) (string, error) {
	replacer := replacement.GetReplacer(ref.Path)
	if replacer == nil {
		return "", fmt.Errorf("unknown file type, unable to replace references")
	}
	val, err := replacer.Retrieve(contents, ref.Field)
	if err != nil {
		return "", fmt.Errorf("cannot read reference file: %s, %v", ref.Path, err)
	}
	replacement, err := getReplacementValue(image, val, ref.Template)
	if err != nil {
		return "", fmt.Errorf("error replacing image")
	}
	return replacer.Replace(contents, ref.Field, replacement)

}

func getReplacementValue(image string, value string, templateVal string) (string, error) {
	imageRef, err := getDockerReference(image)
	if err != nil {
		return "", err
	}
	if templateVal != "" {
		var replacement bytes.Buffer
		goTemplate, err := template.New("").Option("missingkey=error").Parse(templateVal)
		if err != nil {
			return "", fmt.Errorf("error building template: %s", err)
		}
		if err = goTemplate.Execute(&replacement, imageRef); err != nil {
			return "", fmt.Errorf("error executing template: %s", err)
		}
		return replacement.String(), nil
	}
	// This is the fallback behavior if no template is specified
	// If a tag is not present, just include the repo / image portion of the reference
	r, err := getDockerReference(value)
	if err != nil {
		return "", fmt.Errorf("invalid image reference :%s", err)
	}
	if r.Tag == "" {
		return strings.Join([]string{imageRef.Repo, imageRef.Image}, "/"), nil
	}
	return image, nil
}

func (r *Rewriter) loadReferenceFile(path string) (string, os.FileInfo, error) {
	fi, err := r.Context.FileSystem.Stat(path)
	if err != nil {
		return "", nil, err
	}
	refFile, err := r.Context.FileSystem.Open(path)
	defer refFile.Close()
	if err != nil {
		return "", nil, fmt.Errorf("cannot open reference file: %v", err)
	}
	b, err := ioutil.ReadAll(refFile)
	if err != nil {
		return "", nil, fmt.Errorf("cannot read reference file: %v", err)
	}
	return string(b), fi, nil
}

func (r *Rewriter) writeReferenceFile(contents []byte, fi os.FileInfo, path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fi.Mode())
	defer f.Close()
	if err != nil {
		return fmt.Errorf("cannot write reference file: %v", err)
	}
	_, err = f.Write(contents)
	if err != nil {
		return fmt.Errorf("cannot write referenc file: %v", err)
	}
	return nil
}
