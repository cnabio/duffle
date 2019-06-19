package remote

import (
	"fmt"
	"io"

	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"

	"github.com/deislabs/duffle/pkg/imagestore"
)

type Builder struct {
	registryClient registry.Client
}

// remote is an image store which does not actually store images. It is used to represent thin bundles.
type remote struct {
	registryClient registry.Client
}

func NewRemoteBuilder() *Builder {
	return &Builder{
		registryClient: registry.NewRegistryClient(),
	}
}

func (b *Builder) ArchiveDir(archiveDir string) imagestore.Builder {
	return b
}

func (b *Builder) Logs(logs io.Writer) imagestore.Builder {
	return b
}

func (b *Builder) Build() (imagestore.Store, error) {
	return &remote{
		registryClient: b.registryClient,
	}, nil
}

func (r *remote) Add(im string) (string, error) {
	return "", nil
}

func (r *remote) Push(d image.Digest, src image.Name, dst image.Name) error {
	dig, _, err := r.registryClient.Copy(src, dst)
	if err != nil {
		return err
	}

	if d != image.EmptyDigest && dig != d {
		return fmt.Errorf("digest of image %s not preserved: old digest %s; new digest %s", src, d.String(), dig.String())
	}
	return nil
}
