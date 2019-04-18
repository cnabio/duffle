package packager

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// tarFiles is an ImageStore which stores images as separate tar files.
type tarFiles struct {
	Client  *client.Client
	Context context.Context

	artifactsDir string
	logs         io.Writer
}

func newTarFiles() (*tarFiles, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	cli.NegotiateAPIVersion(ctx)
	return &tarFiles{
		Client:  cli,
		Context: ctx,
	}, nil

}

func (t *tarFiles) configure(archiveDir string, logs io.Writer) error {
	artifactsDir := filepath.Join(archiveDir, "artifacts")
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return err
	}
	t.artifactsDir = artifactsDir
	t.logs = logs
	return nil
}

func (t *tarFiles) add(image string) (string, error) {
	ctx := t.Context

	imagePullOptions := types.ImagePullOptions{} //TODO: add platform info
	pullLogs, err := t.Client.ImagePull(ctx, image, imagePullOptions)
	if err != nil {
		return "", fmt.Errorf("Error pulling image %s: %s", image, err)
	}
	defer pullLogs.Close()
	io.Copy(t.logs, pullLogs)

	di, err := t.Client.DistributionInspect(ctx, image, "")
	if err != nil {
		return "", fmt.Errorf("Error inspecting image %s: %s", image, err)
	}

	reader, err := t.Client.ImageSave(ctx, []string{image})
	if err != nil {
		return "", err
	}
	defer reader.Close()

	name := buildFileName(image) + ".tar"
	out, err := os.Create(filepath.Join(t.artifactsDir, name))
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	if err != nil {
		return "", err
	}

	return string(di.Descriptor.Digest), nil
}

func buildFileName(uri string) string {
	filename := strings.Replace(uri, "/", "-", -1)
	return strings.Replace(filename, ":", "-", -1)

}
