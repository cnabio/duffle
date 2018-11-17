package packager

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"

	"github.com/deis/duffle/pkg/loader"
)

var (
	// ErrNoArtifactsDirectory indicates a missing artifacts/ directory
	ErrNoArtifactsDirectory = errors.New("No artifacts/ directory found")
)

// Importer is responsible for importing a file
type Importer struct {
	Source      string
	Destination string
	Client      *client.Client
	Loader      loader.Loader
}

// NewImporter creates a new secure *Importer
//
// source is the filesystem path to the archive.
// destination is the directory to unpack the contents.
// load is a loader.Loader preconfigured for loading secure or insecure bundles.
func NewImporter(source, destination string, load loader.Loader) (*Importer, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	cli.NegotiateAPIVersion(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cannot negotation Docker client version: %v", err)
	}

	return &Importer{
		Source:      source,
		Destination: destination,
		Client:      cli,
		Loader:      load,
	}, nil
}

// Import decompresses a bundle from Source (location of the compressed bundle) and properly places artifacts in the correct location(s)
func (im *Importer) Import() error {
	tempDir, err := ioutil.TempDir("", "duffle-import")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	reader, err := os.Open(im.Source)
	if err != nil {
		return err
	}
	defer reader.Close()

	tarOptions := &archive.TarOptions{
		Compression:      archive.Gzip,
		IncludeFiles:     []string{"."},
		IncludeSourceDir: true,
		// Issue #416
		NoLchown: true,
	}
	if err := archive.Untar(reader, tempDir, tarOptions); err != nil {
		return fmt.Errorf("untar failed: %s", err)
	}

	// We try to load a bundle.cnab file first, and fall back to a bundle.json
	ext := "cnab"
	if _, err := os.Stat(filepath.Join(im.Destination, "bundle.cnab")); os.IsNotExist(err) {
		ext = "json"
	}

	bun, err := im.Loader.Load(filepath.Join(tempDir, "bundle."+ext))
	if err != nil {
		return fmt.Errorf("failed to load bundle.%s: %s", ext, err)
	}

	bunDir := filepath.Join(im.Destination, bun.Name)
	if _, err := os.Stat(bunDir); os.IsNotExist(err) {
		if err := os.Rename(tempDir, bunDir); err != nil {
			return fmt.Errorf("move failed: %s", err)
		}
	} else {
		return fmt.Errorf("Attempted to unpack bundle to %s but path already exists", bunDir)
	}

	artifactsDir := filepath.Join(bunDir, "artifacts")
	_, err = os.Stat(artifactsDir)
	if err == nil {
		filepath.Walk(artifactsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = im.Client.ImageLoad(context.Background(), file, true) // quiet = true
			return err
		})
	}

	return nil
}
