package builder

import (
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"

	"github.com/cnabio/duffle/pkg/duffle/manifest"
	"github.com/cnabio/duffle/pkg/imagebuilder"
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
	outputs := map[string]bundle.Output{"output1": {}}
	params := map[string]bundle.Parameter{"param1": {}}

	mfst := &manifest.Manifest{
		Actions:     map[string]bundle.Action{"act1": {}},
		Credentials: map[string]bundle.Credential{"cred1": {}},
		Custom:      map[string]interface{}{"cus1": nil},
		Definitions: map[string]*definition.Schema{"def1": {}},
		Description: "description",
		Images:      map[string]bundle.Image{"img1": {}},
		InvocationImages: map[string]*manifest.InvocationImage{
			"cnab": {
				Name:          "cnab",
				Configuration: map[string]string{"registry": "registry"},
			},
		},
		Keywords: []string{"test"},
		Maintainers: []bundle.Maintainer{
			{
				Name:  "test",
				Email: "test@test.com",
				URL:   "https://test.com",
			},
		},
		Name:               "foo",
		Outputs:            outputs,
		Parameters:         params,
		SchemaVersion:      "v1.0.0",
		Version:            "0.1.0",
		License:            "MIT",
		RequiredExtensions: []string{"ext1", "ext2"},
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
	checksPerformed := 0

	if !reflect.DeepEqual(b.Actions, mfst.Actions) {
		t.Errorf("expected actions to be %+v but was %+v", mfst.Actions, b.Actions)
	}
	checksPerformed++

	if !reflect.DeepEqual(b.Credentials, mfst.Credentials) {
		t.Errorf("expected credentials to be %+v but was %+v", mfst.Credentials, b.Credentials)
	}
	checksPerformed++

	if !reflect.DeepEqual(b.Custom, mfst.Custom) {
		t.Errorf("expected custom to be %+v but was %+v", mfst.Custom, b.Custom)
	}
	checksPerformed++

	if !reflect.DeepEqual(b.Definitions, mfst.Definitions) {
		t.Errorf("expected definitions to be %+v but was %+v", mfst.Definitions, b.Definitions)
	}
	checksPerformed++

	if b.Description != mfst.Description {
		t.Errorf("expected description to be %+v but was %+v", mfst.Description, b.Description)
	}
	checksPerformed++

	if len(b.InvocationImages) != 1 {
		t.Fatalf("expected there to be 1 image, got %d. Full output: %v", len(b.Images), b)
	}
	checksPerformed++

	expected := bundle.InvocationImage{}
	expected.Image = "cnab:0.1.0"
	expected.ImageType = "docker"
	if !reflect.DeepEqual(b.InvocationImages[0], expected) {
		t.Errorf("expected %v, got %v", expected, b.InvocationImages[0])
	}
	checksPerformed++

	if !reflect.DeepEqual(b.Keywords, mfst.Keywords) {
		t.Errorf("expected keywords to be %+v but was %+v", mfst.Keywords, b.Keywords)
	}
	checksPerformed++

	if !reflect.DeepEqual(b.Maintainers, mfst.Maintainers) {
		t.Errorf("expected maintainers to be %+v but was %+v", mfst.Maintainers, b.Maintainers)
	}
	checksPerformed++

	if b.Name != mfst.Name {
		t.Errorf("expected name to be %+v but was %+v", mfst.Name, b.Name)
	}
	checksPerformed++

	if !reflect.DeepEqual(b.Outputs, mfst.Outputs) {
		t.Errorf("expected outputs to be %+v but was %+v", mfst.Outputs, b.Outputs)
	}
	checksPerformed++

	if !reflect.DeepEqual(b.Parameters, mfst.Parameters) {
		t.Errorf("expected parameters to be %+v but was %+v", mfst.Parameters, b.Parameters)
	}
	checksPerformed++

	if b.SchemaVersion != mfst.SchemaVersion {
		t.Errorf("expected schemaVersion %v, got %v", mfst.SchemaVersion, b.SchemaVersion)
	}
	checksPerformed++

	if b.Version != mfst.Version {
		t.Errorf("expected version %v, got %v", mfst.Version, b.Version)
	}
	checksPerformed++

	if b.License != mfst.License {
		t.Errorf("expected licnse %v, got %v", mfst.License, b.License)
	}
	checksPerformed++

	if !reflect.DeepEqual(b.RequiredExtensions, mfst.RequiredExtensions) {
		t.Errorf("expected credentials to be %+v but was %+v", mfst.RequiredExtensions, b.RequiredExtensions)
	}
	checksPerformed++

	// Ensure that all the fields have been checked. If the structures need to diverge in the future, this test should be modified.
	mfstFields := getFields(manifest.Manifest{})
	if len(mfstFields) != checksPerformed {
		t.Errorf("expected to check %v fields for equality, but checked only %v fields", len(mfstFields), checksPerformed)
	}
}

func TestBundleAndManifestHaveSameFields(t *testing.T) {
	mfst := manifest.Manifest{}
	mfstFields := getFields(mfst)

	b := bundle.Bundle{}
	bundleFields := getFields(b)

	if !reflect.DeepEqual(bundleFields, mfstFields) {
		t.Errorf("manifest and bundle have different fields.\nmanifest: %+v\nbundle: %+v\n", mfstFields, bundleFields)
	}
}

func getFields(i interface{}) map[string]struct{} {
	fields := make(map[string]struct{}, 15)

	v := reflect.ValueOf(i)
	typeOf := reflect.TypeOf(i)

	for i := 0; i < v.NumField(); i++ {
		fields[typeOf.Field(i).Name] = struct{}{}
	}
	return fields
}
