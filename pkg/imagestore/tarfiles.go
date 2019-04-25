package imagestore

import (
	"context"
	"fmt"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/registry"
	"github.com/pivotal/image-relocation/pkg/image"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type tarFilesBuilder struct {
	client     *client.Client
	context    context.Context
	archiveDir string
	logs       io.Writer
}

// tarFiles is an image store which stores images as separate tar files.
type tarFiles struct {
	client       *client.Client
	context      context.Context
	artifactsDir string
	logs         io.Writer
}

func newTarFilesBuilder() (*tarFilesBuilder, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	cli.NegotiateAPIVersion(ctx)
	return &tarFilesBuilder{
		client:     cli,
		context:    ctx,
		archiveDir: "",
		logs:       ioutil.Discard,
	}, nil

}

func (b *tarFilesBuilder) ArchiveDir(archiveDir string) Builder {
	return &tarFilesBuilder{
		client:     b.client,
		context:    b.context,
		archiveDir: archiveDir,
		logs:       b.logs,
	}
}

func (b *tarFilesBuilder) Logs(logs io.Writer) Builder {
	return &tarFilesBuilder{
		client:     b.client,
		context:    b.context,
		archiveDir: b.archiveDir,
		logs:       logs,
	}
}

func (b *tarFilesBuilder) Build() (Store, error) {
	artifactsDir := filepath.Join(b.archiveDir, "artifacts")
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return nil, err
	}
	return &tarFiles{
		client:       b.client,
		context:      b.context,
		artifactsDir: artifactsDir,
		logs:         b.logs,
	}, nil
}

func (t *tarFiles) Add(image string) (string, error) {
	ctx := t.context

	imagePullOptions := types.ImagePullOptions{} //TODO: add platform info
	pullLogs, err := t.client.ImagePull(ctx, image, imagePullOptions)
	if err != nil {
		return "", fmt.Errorf("Error pulling image %s: %v", image, err)
	}
	defer pullLogs.Close()
	io.Copy(t.logs, pullLogs)

	di, err := t.client.DistributionInspect(ctx, image, "")
	if err != nil {
		return "", fmt.Errorf("Error inspecting image %s: %v", image, err)
	}

	reader, err := t.client.ImageSave(ctx, []string{image})
	if err != nil {
		return "", fmt.Errorf("Error saving image %s: %v", image, err)
	}
	defer reader.Close()

	out, err := os.Create(t.imagePath(image))
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

func (t *tarFiles) Push(dig image.Digest, src image.Name, dst image.Name) error {
	in, err := os.Open(t.imagePath(src.String()))
	if err != nil {
		// Accommodate images named using a synonym.
		// See https://github.com/deislabs/cnab-spec/issues/171
		for _, syn := range src.Synonyms() {
			in, err = os.Open(t.imagePath(syn.String()))
			if err == nil {
				break
			}
		}
		if err != nil {
			return err
		}
	}
	resp, err := t.client.ImageLoad(t.context, in, false)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("Error loading image %s: %v", src, err)
	}

	err = t.client.ImageTag(t.context, src.String(), dst.WithoutDigest().String())
	if err != nil {
		return fmt.Errorf("Error tagging image %s as %s: %v", src, dst, err)
	}

	ref, err := reference.ParseNormalizedNamed(dst.String())
	if err != nil {
		return err
	}

	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return err
	}

	cli, err := command.NewDockerCli()
	if err != nil {
		return err
	}
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return err
	}

	authConfig := command.ResolveAuthConfig(t.context, cli, repoInfo.Index)
	encodedAuth, err := command.EncodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}

	imagePushOptions := types.ImagePushOptions{
		RegistryAuth: encodedAuth,
	} //TODO: add platform info
	pushResp, err := t.client.ImagePush(t.context, dst.WithoutDigest().String(), imagePushOptions)
	if err != nil {
		return fmt.Errorf("Error pushing image %s: %v", dst, err)
	}
	defer pushResp.Close()

	// If a digest was specified, check that it was preserved.
	if dig.String() != "" {
		di, err := t.client.DistributionInspect(t.context, dst.String(), encodedAuth)
		if err != nil {
			return fmt.Errorf("Error inspecting image %s: %v", dst, err)
		}

		newDig := di.Descriptor.Digest.String()
		if newDig != dig.String() {
			return fmt.Errorf("Digest modified when pushing %s to %s: old digest %s, new digest %s", src, dst, dig.String(), newDig)
		}
	}

	return err
}

func (t *tarFiles) imagePath(image string) string {
	name := buildFileName(image) + ".tar"
	path := filepath.Join(t.artifactsDir, name)
	return path
}

func buildFileName(uri string) string {
	filename := strings.Replace(uri, "/", "-", -1)
	return strings.Replace(filename, ":", "-", -1)

}
