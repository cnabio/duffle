package packager

import (
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
	"github.com/docker/docker/pkg/archive"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/loader"
)

var (
	ErrDestinationNotDirectory = errors.New("Destination not directory")
)

type Exporter struct {
	Source      string
	Destination string
	Full        bool
	Client      *client.Client
	Context     context.Context
}

// NewExporter returns an *Exporter given information about where a bundle
//  lives, where the compressed bundle should be exported to,
//  and what form a bundle should be exported in (thin or thick/full). It also
//  sets up a docker client to work with images.
func NewExporter(source, dest string, full bool) (*Exporter, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	cli.NegotiateAPIVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot negotation Docker client version: %v", err)
	}

	return &Exporter{
		Source:      source,
		Destination: dest,
		Full:        full,
		Client:      cli,
		Context:     ctx,
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
	writer, err := os.Create(filepath.Join(ex.Destination, name+".tgz"))
	if err != nil {
		return err
	}
	defer writer.Close()

	tarOptions := &archive.TarOptions{
		Compression:      archive.Gzip,
		IncludeFiles:     []string{"."},
		IncludeSourceDir: true,
	}
	rc, err := archive.TarWithOptions(ex.Source, tarOptions)
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.Copy(writer, rc)
	return err
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
	ctx := ex.Context

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
