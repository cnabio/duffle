package digester

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
)

func TestDigest(t *testing.T) {
	d := Digester{
		Client: &mockDockerClient{},
		Image:  "mock-image",
	}
	digestStr, err := d.Digest()
	if err != nil {
		t.Errorf("Not expecting error computing digest but got error: %s", err)
	}
	expectedDigestStr := "sha256:5dffd8ab8b1b8db6fbb2a62f5059d930fe79f3e3d7b2e65d331af5b8de03c93c"
	if digestStr != expectedDigestStr {
		t.Errorf("Expected digest %s, got %s", expectedDigestStr, digestStr)
	}
}

type mockDockerClient struct{}

func (m *mockDockerClient) ImageSave(ctx context.Context, imageIDs []string) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader([]byte("mock-context-for-digest"))), nil
}

func (m *mockDockerClient) ImageBuild(ctx context.Context, context io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	return types.ImageBuildResponse{}, nil
}
func (m *mockDockerClient) BuildCachePrune(ctx context.Context, opts types.BuildCachePruneOptions) (*types.BuildCachePruneReport, error) {
	return nil, nil
}
func (m *mockDockerClient) BuildCancel(ctx context.Context, id string) error {
	return nil
}

func (m *mockDockerClient) ImageCreate(ctx context.Context, parentReference string, options types.ImageCreateOptions) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockDockerClient) ImageHistory(ctx context.Context, image string) ([]image.HistoryResponseItem, error) {
	return nil, nil
}
func (m *mockDockerClient) ImageImport(ctx context.Context, source types.ImageImportSource, ref string, options types.ImageImportOptions) (io.ReadCloser, error) {
	return nil, nil
}
func (m *mockDockerClient) ImageInspectWithRaw(ctx context.Context, image string) (types.ImageInspect, []byte, error) {
	return types.ImageInspect{}, nil, nil
}
func (m *mockDockerClient) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	return nil, nil
}
func (m *mockDockerClient) ImageLoad(ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
	return types.ImageLoadResponse{}, nil
}
func (m *mockDockerClient) ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
	return nil, nil
}
func (m *mockDockerClient) ImagePush(ctx context.Context, ref string, options types.ImagePushOptions) (io.ReadCloser, error) {
	return nil, nil
}
func (m *mockDockerClient) ImageRemove(ctx context.Context, image string, options types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error) {
	return nil, nil
}
func (m *mockDockerClient) ImageSearch(ctx context.Context, term string, options types.ImageSearchOptions) ([]registry.SearchResult, error) {
	return nil, nil
}
func (m *mockDockerClient) ImageTag(ctx context.Context, image, ref string) error {
	return nil
}
func (m *mockDockerClient) ImagesPrune(ctx context.Context, pruneFilter filters.Args) (types.ImagesPruneReport, error) {
	return types.ImagesPruneReport{}, nil
}
