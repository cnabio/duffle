package packager

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
)

var (
	ErrNoArtifactsDirectory = errors.New("No artifacts/ directory found")
)

type Importer struct {
	Source      string
	Destination string
	Client      *client.Client
}

// NewImporter takes a source and destination and returns an *Importer
func NewImporter(source, destination string) (*Importer, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion(dockerClientVersion))
	if err != nil {
		return nil, err
	}

	return &Importer{
		Source:      source,
		Destination: destination,
		Client:      cli,
	}, nil
}

// Import decompresses a bundle from Source (location of the compressed bundle) and properly places artifacts in the correct location(s)
func (im *Importer) Import() error {
	ctx := context.Background()

	reader, err := os.Open(im.Source)
	if err != nil {
		return err
	}
	defer reader.Close()

	gzipReader, err := gzip.NewReader(reader)
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	root := ""
	rootRead := false

	for {
		exit := false
		header, err := tarReader.Next()

		switch {
		case err == io.EOF:
			exit = true
		case err != nil:
			return err
		case header == nil:
			continue
		}

		if exit == true {
			break
		}

		target := filepath.Join(im.Destination, header.Name)
		if !rootRead {
			root = target
			rootRead = true
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}

			file.Close()
		}
	}

	artifactsDir := filepath.Join(root, "artifacts")
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

			_, err = im.Client.ImageLoad(ctx, file, true) // quiet = true
			return err
		})
	}

	return nil
}
