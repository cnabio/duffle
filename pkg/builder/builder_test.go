package builder

import (
	"context"
	"io"
	"testing"

	"github.com/deislabs/cnab-go/bundle"

	"github.com/deislabs/duffle/pkg/duffle/manifest"
	"github.com/deislabs/duffle/pkg/imagebuilder"
)

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

	if len(b.InvocationImages) != 0 {
		t.Errorf("Expected 0 invocation images (no info set until images are built), got %v", len(b.InvocationImages))
	}
	if len(bldr.ImageBuilders) != 1 {
		t.Fatalf("Expected 1 image builder to be set, got %v", len(bldr.ImageBuilders))
	}
	ib := bldr.ImageBuilders[0]
	if ib.Name() != "cnab" {
		t.Errorf("Expected name of invocation image to be cnab, got %s", ib.Name())
	}
	if ib.Type() != "docker" {
		t.Errorf("Expected type of invocation image to be docker, got %s", ib.Type())
	}
	if ib.URI() != "cnab:0.1.0" {
		t.Errorf("Expected URI of invocation image to be cnab:0.1.0, got %s", ib.URI())
	}
}

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
func (tc testImage) Build(ctx context.Context, log io.WriteCloser) (string, error) {
	return "", nil
}
