package imagestore

import (
	"fmt"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
	"io"
)

type remoteBuilder struct {
	registryClient registry.Client
}

// remote is an image store which does not actually store images. It is used to represent thin bundles.
type remote struct {
	registryClient registry.Client
}

func newRemoteBuilder() *remoteBuilder {
	return &remoteBuilder{
		registryClient: registry.NewRegistryClient(),
	}
}

func (b *remoteBuilder) ArchiveDir(archiveDir string) Builder {
	if archiveDir != "" {
		panic("non-empty archive dir specified for remote image store")
	}
	return b
}

func (b *remoteBuilder) Logs(logs io.Writer) Builder {
	return b
}

func (b *remoteBuilder) Build() (Store, error) {
	return &remote{
		registryClient: b.registryClient,
	}, nil
}

func (r *remote) Add(im string) (string, error) {
	return "", nil
}

func (r *remote) Push(d image.Digest, src image.Name, dst image.Name) error {
	dig, err := r.registryClient.Copy(src, dst)
	if err != nil {
		return err
	}

	if d != image.EmptyDigest && dig != d {
		return fmt.Errorf("digest of image %s not preserved: old digest %s; new digest %s", src, d.String(), dig.String())
	}
	return nil
}
