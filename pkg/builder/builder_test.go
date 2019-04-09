package builder

import (
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/deislabs/cnab-go/bundle"

	"github.com/deislabs/duffle/pkg/duffle/manifest"
	"github.com/deislabs/duffle/pkg/imagebuilder"
)

// testImage represents a mock invocation image
type testImage struct {
	Nam   string
	Typ   string
	UR    string
	Diges string
}

// Name represents the name of a mock invocation image
func (tc testImage) Name() string {
	return tc.Nam
}

// Type represents the type of a mock invocation image
func (tc testImage) Type() string {
	return tc.Typ
}

// URI represents the URI of the artefact of a mock invocation image
func (tc testImage) URI() string {
	return tc.UR
}

// Digest represents the digest of a mock invocation image
func (tc testImage) Digest() string {
	return tc.Diges
}

// PrepareBuild is no-op for a mock invocation image
func (tc *testImage) PrepareBuild(appDir, registry, name string) error {
	return nil
}

// Build is no-op for a mock invocation image
func (tc testImage) Build(ctx context.Context, log io.WriteCloser) error {
	return nil
}

func TestPrepareBuild(t *testing.T) {
	mfst := &manifest.Manifest{
		Name:        "foo",
		Version:     "0.1.0",
		Description: "description",
		Keywords:    []string{"test"},
		InvocationImages: map[string]*manifest.InvocationImage{
			"cnab": {
				Name:          "cnab",
				Configuration: map[string]string{"registry": "registry"},
			},
		},
		Maintainers: []bundle.Maintainer{
			{
				Name:  "test",
				Email: "test@test.com",
				URL:   "https://test.com",
			},
		},
	}

	components := []imagebuilder.ImageBuilder{
		&testImage{
			Nam:   "cnab",
			Typ:   "docker",
			UR:    "cnab:0.1.0",
			Diges: "",
		},
	}

	bldr := New()
	_, b, err := bldr.PrepareBuild(bldr, mfst, "", components)
	if err != nil {
		t.Error(err)
	}

	if len(b.InvocationImages) != 1 {
		t.Fatalf("expected there to be 1 image, got %d. Full output: %v", len(b.Images), b)
	}

	expected := bundle.InvocationImage{}
	expected.Image = "cnab:0.1.0"
	expected.ImageType = "docker"
	if !reflect.DeepEqual(b.InvocationImages[0], expected) {
		t.Errorf("expected %v, got %v", expected, b.InvocationImages[0])
	}
}
