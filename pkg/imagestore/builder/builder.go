package builder

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/deislabs/duffle/pkg/imagestore"
	"github.com/deislabs/duffle/pkg/imagestore/ocilayout"
	"github.com/deislabs/duffle/pkg/imagestore/remote"
)

// NewBuilder creates an image store builder which will, if necessary, create archive contents.
func NewBuilder(remoteRepos bool) (imagestore.Builder, error) {
	// infer the concrete type of the image store from the input parameters
	if remoteRepos {
		return remote.NewRemoteBuilder(), nil
	}
	return ocilayout.NewOciLayout(), nil
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

	return ocilayout.LocateOciLayout(b.archiveDir)
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
