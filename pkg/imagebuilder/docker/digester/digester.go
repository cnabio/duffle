package digester

import (
	"golang.org/x/net/context"

	"github.com/docker/docker/client"
	"github.com/opencontainers/go-digest"
)

type Digester struct {
	Client  client.ImageAPIClient
	Image   string
	Context context.Context
}

// NewDigester returns a Digester given the args client, image, ctx
//
// ctx is context to use and pass to Docker client
// client allows us to talk to the Docker client
// image is a string identifier of the image we want to compute digest of
func NewDigester(ctx context.Context, client client.ImageAPIClient, image string) *Digester {
	return &Digester{
		Client:  client,
		Image:   image,
		Context: ctx,
	}
}

// Digest returns the digest of the image tar
func (d *Digester) Digest() (string, error) {
	reader, err := d.Client.ImageSave(d.Context, []string{d.Image})
	if err != nil {
		return "", err
	}
	computedDigest, err := digest.Canonical.FromReader(reader)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	return computedDigest.String(), nil
}
