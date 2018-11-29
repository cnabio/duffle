package image

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/registry"
	"github.com/stretchr/testify/assert"
)

type fakeDockerCli struct {
	command.DockerCli
	client fakeDockerClient
}

func (cli fakeDockerCli) Client() client.APIClient {
	return &cli.client
}

func (cli fakeDockerCli) ConfigFile() *configfile.ConfigFile {
	return &configfile.ConfigFile{}
}

type fakeDockerClient struct {
	client.Client
	localImagesDigests  map[string][]string
	pulledImagesDigests map[string][]string
	pushedImagesDigests map[string][]string
}

type notFoundMock struct {
}

func (notFoundMock) Error() string {
	return "Not found"
}
func (notFoundMock) NotFound() bool {
	return true
}

func (c *fakeDockerClient) ImageInspectWithRaw(ctx context.Context, image string) (types.ImageInspect, []byte, error) {
	if digests, ok := c.localImagesDigests[image]; ok {
		return types.ImageInspect{
			RepoDigests: digests,
		}, nil, nil
	}
	return types.ImageInspect{}, nil, notFoundMock{}
}

func (c *fakeDockerClient) Info(ctx context.Context) (types.Info, error) {
	return types.Info{
		IndexServerAddress: registry.IndexServer,
	}, nil
}

func (c *fakeDockerClient) ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
	c.localImagesDigests[ref] = c.pulledImagesDigests[ref]
	return ioutil.NopCloser(bytes.NewBuffer(nil)), nil
}

func (c *fakeDockerClient) ImagePush(ctx context.Context, ref string, options types.ImagePushOptions) (io.ReadCloser, error) {
	c.localImagesDigests[ref] = c.pushedImagesDigests[ref]
	return ioutil.NopCloser(bytes.NewBuffer(nil)), nil
}

func TestResolveImage(t *testing.T) {
	is := assert.New(t)
	testee := &Resolver{dockerCli: &fakeDockerCli{}}
	resolvedImage, resolvedDigest, err := testee.Resolve("test-image", "sha256:test-digest")
	is.NoError(err)
	is.Equal("test-image@sha256:test-digest", resolvedImage)
	is.Equal("sha256:test-digest", resolvedDigest)

	testee.dockerCli = &fakeDockerCli{
		DockerCli: *command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false),
		client: fakeDockerClient{
			localImagesDigests: map[string][]string{
				"test-image": {"test-image@sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341"},
			},
		},
	}
	resolvedImage, resolvedDigest, err = testee.Resolve("test-image", "")
	is.NoError(err)
	is.Equal("test-image@sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341", resolvedImage)
	is.Equal("sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341", resolvedDigest)
	testee.dockerCli = &fakeDockerCli{
		DockerCli: *command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false),
		client: fakeDockerClient{
			localImagesDigests: map[string][]string{},
			pulledImagesDigests: map[string][]string{
				"test-image": {"test-image@sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341"},
			},
		},
	}
	resolvedImage, resolvedDigest, err = testee.Resolve("test-image", "")
	is.NoError(err)
	is.Equal("test-image@sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341", resolvedImage)
	is.Equal("sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341", resolvedDigest)

	testee.dockerCli = &fakeDockerCli{
		DockerCli: *command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false),
		client: fakeDockerClient{
			localImagesDigests: map[string][]string{
				"test-image": nil,
			},
			pushedImagesDigests: map[string][]string{
				"test-image": {"test-image@sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341"},
			},
		},
	}
	_, _, err = testee.Resolve("test-image", "")
	isLocalOnly, image := IsErrImageLocalOnly(err)
	is.True(isLocalOnly)
	is.Equal(image, "test-image")

	testee.pushLocalImages = true
	resolvedImage, resolvedDigest, err = testee.Resolve("test-image", "")
	is.NoError(err)
	is.Equal("test-image@sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341", resolvedImage)
	is.Equal("sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341", resolvedDigest)

	testee.dockerCli = &fakeDockerCli{
		DockerCli: *command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false),
		client: fakeDockerClient{
			localImagesDigests: map[string][]string{
				"test-image": {"other-registry:5000/namespace/test-image@sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341"},
			},
			pushedImagesDigests: map[string][]string{
				"test-image": {"test-image@sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341"},
			},
		},
	}
	resolvedImage, resolvedDigest, err = testee.Resolve("test-image", "")
	is.NoError(err)
	is.Equal("test-image@sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341", resolvedImage)
	is.Equal("sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341", resolvedDigest)
}
