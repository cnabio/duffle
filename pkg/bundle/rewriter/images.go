package rewriter

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// This interface makes it easier to mock the docker client for testing this
type dockerImageClient interface {
	ImageInspectWithRaw(ctx context.Context, image string) (types.ImageInspect, []byte, error)
	ImageTag(ctx context.Context, image, ref string) error
}

// Only using the image api client
func getDockerClient() (dockerImageClient, error) {
	return client.NewClientWithOpts(client.WithVersion("1.37"))
}

// ReplaceRepository will replace the repository portion of a docker image
func (r *rewriter) ReplaceRepository(qualifiedImage string, repository string) (string, error) {
	dr, err := getDockerReference(qualifiedImage)
	if err != nil {
		return "", err
	}
	if dr.Tag != "" {
		return fmt.Sprintf("%s/%s:%s", repository, dr.Image, dr.Tag), nil
	}
	return fmt.Sprintf("%s/%s", repository, dr.Image), nil
}

// TagImage will locally re-tag a Docker image
func (r *rewriter) TagImage(ctx context.Context, oldImage string, newImage string) error {
	_, _, err := r.dockerClient.ImageInspectWithRaw(ctx, oldImage)
	if err != nil {
		return err
	}
	return r.dockerClient.ImageTag(ctx, oldImage, newImage)
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
