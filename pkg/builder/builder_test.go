package builder

import (
	"context"
	"reflect"
	"testing"

	"github.com/deislabs/duffle/pkg/bundle"

	"github.com/deislabs/duffle/pkg/duffle/manifest"
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
func (tc *testImage) PrepareBuild(ctx *Context) error {
	return nil
}

// Build is no-op for a mock invocation image
func (tc testImage) Build(ctx context.Context, app *AppContext) error {
	return nil
}

func TestPrepareBuild(t *testing.T) {
	mfst := &manifest.Manifest{
		Name:        "foo",
		Version:     "0.1.0",
		Description: "description",
		Keywords:    []string{"test"},
		Maintainers: []bundle.Maintainer{
			{
				Name:  "test",
				Email: "test@test.com",
				URL:   "https://test.com",
			},
		},
	}

	components := []Component{
		&testImage{
			Nam:   "cnab",
			Typ:   "docker",
			UR:    "cnab:0.1.0",
			Diges: "",
		},
		&testImage{
			Nam:   "component1",
			Typ:   "docker",
			UR:    "component1:0.1.0",
			Diges: "",
		},
	}

	bldr := New()
	_, b, err := bldr.PrepareBuild(bldr, mfst, "", components)
	if err != nil {
		t.Error(err)
	}

	if len(b.Images) != 1 {
		t.Fatalf("expected there to be 1 image, got %d. Full output: %v", len(b.Images), b)
	}

	expected := bundle.Image{Description: "component1"}
	expected.Image = "component1:0.1.0"
	expected.ImageType = "docker"
	if !reflect.DeepEqual(b.Images[0], expected) {
		t.Errorf("expected %v, got %v", expected, b.Images[0])
	}
}
