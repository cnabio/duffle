package rewriter

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReplaceRepositoryValidImageWithTag(t *testing.T) {
	oldImage := "mcr.microsoft.com/testinvocationimage:v1.1.0"
	newRepo := "repo/testing"
	expected := "repo/testing/testinvocationimage:v1.1.0"
	is := assert.New(t)
	r := rewriter{}
	newImage, err := r.ReplaceRepository(oldImage, newRepo)
	is.Nil(err)
	is.Equal(expected, newImage)
}

func TestReplaceRepositoryValidImageWithoutTag(t *testing.T) {
	oldImage := "mcr.microsoft.com/testinvocationimage"
	newRepo := "repo/testing"
	expected := "repo/testing/testinvocationimage"
	is := assert.New(t)
	r := rewriter{}
	newImage, err := r.ReplaceRepository(oldImage, newRepo)
	is.Nil(err)
	is.Equal(expected, newImage)
}

func TestReplaceRepositoryValidImageWithoutRepo(t *testing.T) {
	oldImage := "testinvocationimage"
	newRepo := "repo/testing"
	expected := "repo/testing/testinvocationimage"
	is := assert.New(t)
	r := rewriter{}
	newImage, err := r.ReplaceRepository(oldImage, newRepo)
	is.Nil(err)
	is.Equal(expected, newImage)
}

func TestReplaceRepositoryValidImageWithHostAndPort(t *testing.T) {
	oldImage := "localhost:5000/testinvocationimage"
	newRepo := "repo/testing"
	expected := "repo/testing/testinvocationimage"
	is := assert.New(t)
	r := rewriter{}
	newImage, err := r.ReplaceRepository(oldImage, newRepo)
	is.Nil(err)
	is.Equal(expected, newImage)
}

func TestReplaceRepositoryInvalidImage(t *testing.T) {
	oldImage := "mcr.microsoft.com/testinvocationimage:test:test"
	newRepo := "repo/testing"
	is := assert.New(t)
	r := rewriter{}
	_, err := r.ReplaceRepository(oldImage, newRepo)
	is.NotNil(err)
	is.Equal(err, reference.ErrReferenceInvalidFormat)
}

func TestTagImage(t *testing.T) {
	d := new(mockDocker)
	ctx := context.Background()

	oldImage := "testinvocationimage"
	newRepo := "repo/testing"
	expected := "repo/testing/testinvocationimage"

	d.On("ImageInspectWithRaw", ctx, "testinvocationimage").Return(nil, nil, nil)
	d.On("ImageTag", ctx, "testinvocationimage", expected).Return(nil)

	is := assert.New(t)
	r := rewriter{
		dockerClient: d,
	}
	newImage, err := r.ReplaceRepository(oldImage, newRepo)
	is.Nil(err)
	is.Equal(expected, newImage)
	err = r.TagImage(ctx, oldImage, newImage)
	is.Nil(err)
	is.Equal(expected, newImage)
}

func TestMissingImage(t *testing.T) {
	d := new(mockDocker)
	ctx := context.Background()

	oldImage := "testinvocationimage"
	newRepo := "repo/testing"
	err := fmt.Errorf("missing image")
	d.On("ImageInspectWithRaw", ctx, "testinvocationimage").Return(nil, nil, err)
	is := assert.New(t)
	r := rewriter{
		dockerClient: d,
	}
	expectedError := r.TagImage(ctx, oldImage, newRepo)
	is.NotNil(expectedError)
	is.Equal(err, expectedError)
}

type mockDocker struct {
	mock.Mock
}

func (d *mockDocker) ImageInspectWithRaw(ctx context.Context, image string) (types.ImageInspect, []byte, error) {
	args := d.Called(ctx, image)
	inspectResult, _ := args.Get(0).(types.ImageInspect)
	bytes, _ := args.Get(1).([]byte)
	return inspectResult, bytes, args.Error(2)
}

func (d *mockDocker) ImageTag(ctx context.Context, image, ref string) error {
	args := d.Called(ctx, image, ref)
	return args.Error(0)
}
