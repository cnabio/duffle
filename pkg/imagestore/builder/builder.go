package builder

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/deislabs/duffle/pkg/imagestore"
	"github.com/deislabs/duffle/pkg/imagestore/ocilayout"
	"github.com/deislabs/duffle/pkg/imagestore/remote"
	"github.com/deislabs/duffle/pkg/imagestore/tarfiles"
)

// NewBuilder creates an image store builder which will, if necessary, create archive contents.
func NewBuilder(remoteRepos bool, ociLayout bool) (imagestore.Builder, error) {
	// infer the concrete type of the image store from the input parameters
	if remoteRepos {
		return remote.NewRemoteBuilder(), nil
	}
	if ociLayout {
		return ocilayout.NewOciLayout(), nil
	}
	return tarfiles.NewTarFilesBuilder()
}

// NewLocatingBuilder creates an image store builder which will, if necessary, find existing archive contents.
func NewLocatingBuilder() imagestore.Builder {
	return &locatingBuilder{
		archiveDir: "",
		logs:       ioutil.Discard,
	}
}

type locatingBuilder struct {
	archiveDir string
	logs       io.Writer
}

func (b *locatingBuilder) ArchiveDir(archiveDir string) imagestore.Builder {
	return &locatingBuilder{
		archiveDir: archiveDir,
		logs:       b.logs,
	}
}

func (b *locatingBuilder) Logs(logs io.Writer) imagestore.Builder {
	return &locatingBuilder{
		archiveDir: b.archiveDir,
		logs:       logs,
	}
}

func (b *locatingBuilder) Build() (imagestore.Store, error) {
	if thin(b.archiveDir) {
		return remote.NewRemoteBuilder().Build()
	}

	if s, err := ocilayout.LocateOciLayout(b.archiveDir); err == nil {
		return s, nil
	}

	tarFilesBuilder, err := tarfiles.NewTarFilesBuilder()
	if err != nil {
		return nil, err
	}
	return tarFilesBuilder.ArchiveDir(b.archiveDir).Build()
}

func thin(archiveDir string) bool {
	// If there is no archive directory, the bundle is thin
	if archiveDir == "" {
		return true
	}

	// If there is an archive directory, the bundle is thin if and only if the archive directory has no artifacts/
	// subdirectory
	layoutDir := filepath.Join(archiveDir, "artifacts")
	_, err := os.Stat(layoutDir)
	return os.IsNotExist(err)
}
