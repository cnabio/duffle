package packager

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
)

// ociLayout is an ImageStore which stores images as an OCI image layout.
type ociLayout struct {
	registryClient registry.Client

	layout registry.Layout
	logs   io.Writer
}

func newOciLayout() *ociLayout {
	return &ociLayout{
		registryClient: registry.NewRegistryClient(),
	}
}

func (t *ociLayout) configure(archiveDir string, logs io.Writer) error {
	layoutDir := filepath.Join(archiveDir, "artifacts", "layout")
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		return err
	}

	layout, err := t.registryClient.NewLayout(layoutDir)
	if err != nil {
		return err
	}

	t.layout = layout
	t.logs = logs
	return nil
}

func (t *ociLayout) add(im string) (string, error) {
	n, err := image.NewName(im)
	if err != nil {
		return "", err
	}

	dig, err := t.layout.Add(n)
	if err != nil {
		return "", err
	}

	return dig.String(), nil
}
