package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/bundle/rewriter"
	"github.com/spf13/cobra"
)

var (
	bundleFile string
)

func newRewriteCmd(w io.Writer) *cobra.Command {
	const usage = `Rewrite repository references in an existing CNAB bundle

The rewrite command takes a local CNAB bundle.json file and rewrites the invocation
image and any images defined in the images array of the bundle.json file. The command will also
update file references for the images. Images will also be retagged. Docker images must exist locally.  

The repository argument will be used to rewrite the image. 

Example:
	$ duffle rewrite -f hellohelm/bundle.json deis/example

	Running this command against a bundle.json with an invocation image defined as cnab/hellohelm:latest 
	will result in the bundle.json being rewritten to deis/example/hellohelm:latest

After rewriting a bundle, you will need to push the newly tagged images. 
`
	cmd := &cobra.Command{
		Use:   "rewrite",
		Short: "rewrite repository references in an existing bundle",
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := validateArgs(args, bundleFile)
			if err != nil {
				return err
			}

			bundle, err := loadBundle(bundleFile)
			if err != nil {
				return fmt.Errorf("unable to load bundle: %s", err)
			}
			r, err := rewriter.NewRewriter()
			if err != nil {
				return fmt.Errorf("unable to get docker client:%s", err)
			}
			ctx := context.Background()
			err = tagImages(ctx, w, r, &bundle, repo)
			if err != nil {
				return err
			}
			return rewriteFiles(w, r, bundleFile, &bundle)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&bundleFile, "file", "f", "", "bundle file to install")
	return cmd
}

func validateArgs(args []string, bundleFile string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("This command requires a REPOSITORY to use for rewriting.\nValid input:\n\t$ duffle rewrite REPOSITORY -f path-to-bundle.json")
	}
	if bundleFile == "" {
		return "", errors.New("Please specify bundle file with -f path-to-bundle.json.\nValid input:\n\t$ duffle rewrite REPOSITORY -f path-to-bundle.json")
	}
	return args[0], nil
}

func tagImages(ctx context.Context, w io.Writer, rewriter rewriter.Rewriter, bun *bundle.Bundle, repo string) error {

	fmt.Fprintln(w, "Retagging images...")
	newInvocationImage, err := rewriter.ReplaceRepository(bun.InvocationImage.Image, repo)
	if err != nil {
		return err
	}
	err = rewriter.TagImage(ctx, bun.InvocationImage.Image, newInvocationImage)
	if err != nil {
		return fmt.Errorf("unable to tag invocation image: %s", err)
	}
	fmt.Fprintf(w, "retagged '%s' to '%s'\n", bun.InvocationImage.Image, newInvocationImage)
	bun.InvocationImage.Image = newInvocationImage
	var newImage string
	for i, image := range bun.Images {
		newImage, err = rewriter.ReplaceRepository(image.URI, repo)
		if err != nil {
			return fmt.Errorf("unable to update image %s: %s", image.URI, err)
		}
		err := rewriter.TagImage(ctx, image.URI, newImage)
		if err != nil {
			return fmt.Errorf("unable to retag image %s: %s", image.URI, err)
		}
		fmt.Fprintf(w, "retagged '%s' to '%s'\n", image.URI, newImage)
		bun.Images[i].URI = newImage
	}
	return nil
}

func loadReferenceFile(path string) (string, error) {
	refFile, err := os.Open(path)
	defer refFile.Close()
	if err != nil {
		return "", fmt.Errorf("cannot open reference file: %v", err)
	}
	b, err := ioutil.ReadAll(refFile)
	if err != nil {
		return "", fmt.Errorf("cannot read reference file: %v", err)
	}
	return string(b), nil
}

func writeReferenceFile(contents []byte, path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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

func rewriteFiles(w io.Writer, rewriter rewriter.Rewriter, bundleFile string, bun *bundle.Bundle) error {
	dir, _ := filepath.Split(bundleFile)
	fmt.Fprintln(w, "Updating bundle files...")
	for _, image := range bun.Images {
		for _, ref := range image.Refs {
			contents, err := loadReferenceFile(filepath.Join(dir, ref.Path))
			if err != nil {
				return err
			}
			r, err := rewriter.Rewrite(contents, image.URI, ref)
			if err != nil {
				return err
			}
			err = writeReferenceFile([]byte(r), filepath.Join(dir, ref.Path))
			if err != nil {
				return err
			}
			fmt.Fprintln(w, ref.Path)
		}
	}
	f, err := os.OpenFile(bundleFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("cannot open bundle file: %s", err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	if err := enc.Encode(bun); err != nil {
		return fmt.Errorf("cannot write bundle file: %s", err)
	}
	fmt.Fprintln(w, bundleFile)
	return nil
}
