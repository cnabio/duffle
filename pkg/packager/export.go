package packager

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/loader"
)

const (
	dockerClientVersion = "1.38"
	gzipCompression     = "gzip"
)

var (
	ErrDestinationNotDirectory = errors.New("Destination not directory")
)

//type Export interface {
//Export() error
//Compress(string, string) error
//}

type Exporter struct {
	Source      string
	Destination string
	Full        bool
	Client      *client.Client
}

// NewExporter returns an *Exporter given information about where a bundle
//  lives, where the compressed bundle should be exported to,
//  and what form a bundle should be exported in (thin or thick/full). It also
//  sets up a docker client to work with images.
func NewExporter(source, dest string, full bool) (*Exporter, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion(dockerClientVersion))
	if err != nil {
		return nil, err
	}

	return &Exporter{
		Source:      source,
		Destination: dest,
		Full:        full,
		Client:      cli,
	}, nil
}

// Export prepares an artifacts directory containing all of the necessary
//  images, packages the bundle along with the artifacts in a gzipped tar
//  file, and saves that file to the destination
func (ex *Exporter) Export() error {
	l := loader.NewUnsignedLoader() // TODO: switch on flag

	bun, err := l.Load(filepath.Join(ex.Source, "bundle.json"))
	if err != nil {
		return fmt.Errorf("Error loading bundle: %s", err)
	}

	if ex.Full {
		if err := ex.prepareArtifacts(bun); err != nil {
			return fmt.Errorf("Error preparing artifacts: %s", err)
		}
	}

	name := bun.Name + "-" + bun.Version
	tarfile, err := ex.Archive(name)
	if err != nil {
		return err
	}

	return ex.Compress(tarfile, gzipCompression)
}

func (ex *Exporter) Compress(tarfile, strategy string) error {
	if strategy != gzipCompression {
		return fmt.Errorf("%s compression not supported", strategy)
	}

	reader, err := os.Open(tarfile)
	if err != nil {
		return err
	}
	defer reader.Close()
	defer os.Remove(tarfile)

	target := filepath.Base(tarfile) + ".gz"
	name := strings.TrimSuffix(tarfile, ".tar")

	writer, err := os.Create(filepath.Join(ex.Destination, target))
	if err != nil {
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = name
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)
	return err
}

func (ex *Exporter) Archive(name string) (string, error) {
	target := name + ".tar"
	targetPath := filepath.Join(ex.Destination, target)
	tarfile, err := os.Create(targetPath)
	if err != nil {
		return targetPath, err
	}

	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(ex.Source)
	if err != nil {
		return targetPath, err
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(ex.Source)
	} else {
		return targetPath, fmt.Errorf("%s is not a directory", ex.Source)
	}

	return targetPath, filepath.Walk(ex.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, ex.Source))
		if err := tarball.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil // nested directories in artifacts/ not supported
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		defer file.Close()
		_, err = io.Copy(tarball, file)
		return err
	})
}

// PrepareArtifacts pulls all images, verifies their digests (TODO: verify digest) and
//  saves them to a directory called artifacts/ in the bundle directory
func (ex *Exporter) prepareArtifacts(bun *bundle.Bundle) error {
	artifactsDir := filepath.Join(ex.Source, "artifacts")
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return err
	}

	for _, image := range bun.Images {
		_, err := ex.archiveImage(image.URI, artifactsDir)
		if err != nil {
			return err
		}
	}

	for _, in := range bun.InvocationImages {
		_, err := ex.archiveImage(in.Image, artifactsDir)
		if err != nil {
			return err
		}

	}

	return nil
}

func (ex *Exporter) archiveImage(image, artifactsDir string) (string, error) {
	ctx := context.Background()

	imagePullOptions := types.ImagePullOptions{} //TODO: add platform info
	_, err := ex.Client.ImagePull(ctx, image, imagePullOptions)
	if err != nil {
		return "", err
	}
	//TODO: verify digest after pull
	time.Sleep(2 * time.Second) //TODO: get rid of this gah

	reader, err := ex.Client.ImageSave(ctx, []string{image})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	name := buildFileName(image) + ".tar"
	out, err := os.Create(filepath.Join(artifactsDir, name))
	if err != nil {
		return name, err
	}
	defer out.Close()
	if _, err := io.Copy(out, reader); err != nil {
		return name, err
	}

	return name, nil
}

func buildFileName(uri string) string {
	filename := strings.Replace(uri, "/", "-", -1)
	return strings.Replace(filename, ":", "-", -1)

}

func parseImageURI(uri string) (string, string, error) {
	parts := strings.Split(uri, ":")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("Cannot parse tag from Image URI: %v", uri)
	}

	tag := parts[len(parts)-1]
	repo := strings.Join(parts[:len(parts)-1], ":")
	return repo, tag, nil

}
