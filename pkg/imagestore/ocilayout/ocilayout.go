package ocilayout

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"

	"github.com/deislabs/duffle/pkg/imagestore"
)

type Builder struct {
	registryClient registry.Client
	archiveDir     string
	logs           io.Writer
}

// ociLayout is an image store which stores images as an OCI image layout.
type ociLayout struct {
	layout registry.Layout
	logs   io.Writer
}

func NewOciLayout() *Builder {
	return &Builder{
		registryClient: registry.NewRegistryClient(),
		archiveDir:     "",
		logs:           ioutil.Discard,
	}
}

func LocateOciLayout(archiveDir string) (imagestore.Store, error) {
	layoutDir := filepath.Join(archiveDir, "artifacts", "layout")
	if _, err := os.Stat(layoutDir); os.IsNotExist(err) {
		return nil, err
	}
	o := NewOciLayout()
	layout, err := o.registryClient.ReadLayout(layoutDir)
	if err != nil {
		return nil, err
	}

	return &ociLayout{
		layout: layout,
		logs:   ioutil.Discard,
	}, nil
}

func (b *Builder) ArchiveDir(archiveDir string) imagestore.Builder {
	return &Builder{
		registryClient: b.registryClient,
		archiveDir:     archiveDir,
		logs:           b.logs,
	}
}

func (b *Builder) Logs(logs io.Writer) imagestore.Builder {
	return &Builder{
		registryClient: b.registryClient,
		archiveDir:     b.archiveDir,
		logs:           logs,
	}
}

func (b *Builder) Build() (imagestore.Store, error) {
	layoutDir := filepath.Join(b.archiveDir, "artifacts", "layout")
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		return nil, err
	}

	layout, err := b.registryClient.NewLayout(layoutDir)
	if err != nil {
		return nil, err
	}

	return &ociLayout{
		layout: layout,
		logs:   b.logs,
	}, nil

}

func (o *ociLayout) Add(im string) (string, error) {
	n, err := image.NewName(im)
	if err != nil {
		return "", err
	}

	dig, err := o.layout.Add(n)
	if err != nil {
		return "", err
	}

	return dig.String(), nil
}

func (o *ociLayout) Push(dig image.Digest, src image.Name, dst image.Name) error {
	if dig == image.EmptyDigest {
		var err error
		dig, err = o.layout.Find(src)
		if err != nil {
			return err
		}
	}
	return o.layout.Push(dig, dst)
}
