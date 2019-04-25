package imagestore

import (
	"github.com/pivotal/image-relocation/pkg/image"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Builder is a means of creating image stores.
type Builder interface {
	// ArchiveDir creates a fresh Builder with the given archive directory.
	ArchiveDir(string) Builder

	// Logs creates a fresh builder with the given log stream.
	Logs(io.Writer) Builder

	// Build creates an image store.
	Build() (Store, error) // FIXME: differentiate between existing and new contents of file system
}

// Store is an abstract image store.
type Store interface {
	// Add copies the image with the given name to the image store.
	Add(img string) (contentDigest string, err error)

	// Push copies the image with the given digest from an image with the given name in the image store to a repository
	// with the given name.
	Push(dig image.Digest, src image.Name, dst image.Name) error
}

func NewBuilder(remote bool, ociLayout bool) (Builder, error) {
	// infer the concrete type of the image store from the input parameters
	if remote {
		return newRemoteBuilder(), nil
	}
	if ociLayout {
		return newOciLayout(), nil
	}
	return newTarFilesBuilder()
}

type locatingBuilder struct {
	archiveDir string
	logs       io.Writer
}

func (b *locatingBuilder) ArchiveDir(archiveDir string) Builder {
	return &locatingBuilder{
		archiveDir: archiveDir,
		logs:       b.logs,
	}
}

func (b *locatingBuilder) Logs(logs io.Writer) Builder {
	return &locatingBuilder{
		archiveDir: b.archiveDir,
		logs:       logs,
	}
}

func (b *locatingBuilder) Build() (Store, error) {
	if b.archiveDir == "" {
		return newRemoteBuilder().Build()
	}

	layoutDir := filepath.Join(b.archiveDir, "artifacts", "layout") // FIXME: avoid duplication with ociLayout logic
	if _, err := os.Stat(layoutDir); os.IsNotExist(err) {
		tarFilesFactory, err := newTarFilesBuilder()
		if err != nil {
			return nil, err
		}
		return tarFilesFactory.ArchiveDir(b.archiveDir).Build()
	}

	return readOciLayout(layoutDir)

}

func LocatingBuilder() Builder {
	return &locatingBuilder{
		archiveDir: "",
		logs: ioutil.Discard,
	}
}

func Locate(archiveDir string) (Store, error) {
	if archiveDir == "" {
		return newRemoteBuilder().Build()
	}

	layoutDir := filepath.Join(archiveDir, "artifacts", "layout") // FIXME: avoid duplication with ociLayout logic
	if _, err := os.Stat(layoutDir); os.IsNotExist(err) {
		tarFilesFactory, err := newTarFilesBuilder()
		if err != nil {
			return nil, err
		}
		return tarFilesFactory.ArchiveDir(archiveDir).Build()
	}

	return readOciLayout(layoutDir)
}
