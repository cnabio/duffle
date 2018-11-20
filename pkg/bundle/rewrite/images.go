package rewrite

import (
	"archive/tar"
	//	"bufio"
	//	"bytes"
	"context"
	"fmt"
	"io"
	//"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/registry"
)

func (r *Rewriter) getRepoInfo(qualifiedImage string) (*registry.RepositoryInfo, error) {
	ref, err := reference.ParseNormalizedNamed(qualifiedImage)
	if err != nil {
		return nil, err
	}
	return registry.ParseRepositoryInfo(ref)
}

func (r *Rewriter) pullImageIfNeeded(ctx context.Context, qualifiedImage string) error {
	cli := r.Context.DockerClient.Client()
	if cli == nil {
		return fmt.Errorf("unable to obtain docker client")
	}
	_, _, err := r.Context.DockerClient.Client().ImageInspectWithRaw(ctx, qualifiedImage)
	if client.IsErrNotFound(err) {
		repoInfo, err := r.getRepoInfo(qualifiedImage)
		if err != nil {
			return err
		}
		authConfig := command.ResolveAuthConfig(ctx, r.Context.DockerClient, repoInfo.Index)
		encodedAuth, err := command.EncodeAuthToBase64(authConfig)
		if err != nil {
			return err
		}
		options := types.ImagePullOptions{
			RegistryAuth: encodedAuth,
		}
		responseBody, err := r.Context.DockerClient.Client().ImagePull(ctx, qualifiedImage, options)
		if err != nil {
			return err
		}
		defer responseBody.Close()
		return jsonmessage.DisplayJSONMessagesStream(
			responseBody,
			r.Context.DockerClient.Out(),
			r.Context.DockerClient.Out().FD(),
			false, //r.Context.DockerClient.Out().IsTerminal(),
			nil,
		)
	} else if err != nil {
		fmt.Fprintf(r.Writer, "there was an error pulling: %s", err)
		return err
	}
	return nil
}

func (r *Rewriter) createContainer(ctx context.Context, image string) (string, error) {
	cfg := &container.Config{
		Image: image,
	}
	hostCfg := &container.HostConfig{AutoRemove: true}

	resp, err := r.Context.DockerClient.Client().ContainerCreate(ctx, cfg, hostCfg, nil, "")
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (r *Rewriter) getTar(ctx context.Context, containerID string, src string) (io.Reader, error) {
	rdr, wtr := io.Pipe()
	tw := tar.NewWriter(wtr)
	go func() {
		r.Context.FileSystem.Walk(src, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			target := strings.Replace(file, src, "", -1)
			if target == "" {
				return nil
			}
			header, err := tar.FileInfoHeader(fi, fi.Name())
			if err != nil {
				return fmt.Errorf("couldn't make header: %s", err)
			}
			header.Name = strings.TrimPrefix(target, string(filepath.Separator))
			err = tw.WriteHeader(header)
			if err != nil {
				return fmt.Errorf("couldn't write header: %s", err)
			}
			if !fi.Mode().IsRegular() {
				return nil
			}
			f, err := r.Context.FileSystem.Open(file)
			if err != nil {
				return fmt.Errorf("couldn't read file: %s", err)

			}
			if _, err := io.Copy(tw, f); err != nil {
				return fmt.Errorf("couldnt copy file: %s", err)
			}
			return nil
		})
		wtr.Close()
	}()
	return rdr, nil
}

func (r *Rewriter) copyToContainerFromFileSystem(ctx context.Context, containerID string, src string) error {
	rdr, err := r.getTar(ctx, containerID, src)
	if err != nil {
		return err
	}
	return r.Context.DockerClient.Client().CopyToContainer(ctx, containerID, "/", rdr, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
}

func (r *Rewriter) copyFromContainerToFileSystem(ctx context.Context, containerID string, dest string) error {

	rdr, _, err := r.Context.DockerClient.Client().CopyFromContainer(ctx, containerID, "/cnab")
	if err != nil {
		return err
	}
	tr := tar.NewReader(rdr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		path := filepath.Join(dest, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = r.Context.FileSystem.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}
		f, err := r.Context.FileSystem.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, tr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Rewriter) commitContainer(ctx context.Context, containerID string, name string) (string, error) {
	commitOptions := types.ContainerCommitOptions{
		Reference: name,
	}
	resp, err := r.Context.DockerClient.Client().ContainerCommit(ctx, containerID, commitOptions)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (r *Rewriter) stopContainer(ctx context.Context, containerID string) error {
	return r.Context.DockerClient.Client().ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
}

func (r *Rewriter) replaceRepository(image string, repository string) (string, error) {
	dr, err := getDockerReference(image)
	if err != nil {
		return "", err
	}
	if dr.Tag != "" {
		return fmt.Sprintf("%s/%s:%s", repository, dr.Image, dr.Tag), nil
	}
	return fmt.Sprintf("%s/%s", repository, dr.Image), nil

}

func (r *Rewriter) tagImage(ctx context.Context, oldImage string, newImage string) error {
	return r.Context.DockerClient.Client().ImageTag(ctx, oldImage, newImage)
}

func getDockerReference(ref string) (*dockerReference, error) {
	dr := dockerReference{}
	r, err := reference.Parse(ref)
	if err != nil {
		return nil, err
	}
	dr.setImageAndRepo(r)
	dr.setTag(r)
	return &dr, nil
}

type dockerReference struct {
	Repo  string
	Image string
	Tag   string
}

func (d *dockerReference) setImageAndRepo(ref reference.Reference) {
	named, ok := ref.(reference.Named)
	if ok {
		name := named.Name()
		if strings.LastIndex(name, "/") == -1 {
			d.Image = name
		} else {
			d.Repo = name[:strings.LastIndex(name, "/")]
			d.Image = name[strings.LastIndex(name, "/")+1:]
		}
	}
}

func (d *dockerReference) setTag(ref reference.Reference) {
	tagged, ok := ref.(reference.Tagged)
	if ok {
		d.Tag = tagged.Tag()
	}
}
