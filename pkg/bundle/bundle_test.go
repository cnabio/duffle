package bundle

import (
	"io/ioutil"
	"testing"
)

func TestReadTopLevelProperties(t *testing.T) {
	json := `{
		"name": "foo",
		"version": "1.0",
		"images": [],
		"credentials": {}
	}`
	bundle, err := Parse(json)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Name != "foo" {
		t.Errorf("Expected name 'foo', got '%s'", bundle.Name)
	}
	if bundle.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", bundle.Version)
	}
	if len(bundle.Images) != 0 {
		t.Errorf("Expected no images, got %d", len(bundle.Images))
	}
	if len(bundle.Credentials) != 0 {
		t.Errorf("Expected no credentials, got %d", len(bundle.Credentials))
	}
}

func TestReadImageProperties(t *testing.T) {
	data, err := ioutil.ReadFile("../../tests/testdata/home/repositories/github.com/deis/bundles.git/bundles/foo.json")
	if err != nil {
		t.Errorf("cannot read bundle file: %v", err)
	}

	bundle, err := ParseBuffer(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(bundle.Images))
	}
	image1 := bundle.Images[0]
	if image1.Name != "image1" {
		t.Errorf("Expected name 'image1', got '%s'", image1.Name)
	}
	if image1.URI != "urn:image1uri" {
		t.Errorf("Expected URI 'urn:image1uri', got '%s'", image1.URI)
	}
	if len(image1.Refs) != 1 {
		t.Errorf("Expected 1 ref, got %d", len(image1.Refs))
	}
}

func TestReadCredentialProperties(t *testing.T) {
	data, err := ioutil.ReadFile("../../tests/testdata/home/repositories/github.com/deis/bundles.git/bundles/foo.json")
	if err != nil {
		t.Errorf("cannot read bundle file: %v", err)
	}

	bundle, err := ParseBuffer(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Credentials) != 3 {
		t.Errorf("Expected 3 credentials, got %d", len(bundle.Credentials))
	}
	f := bundle.Credentials["foo"]
	if f.Path != "pfoo" {
		t.Errorf("Expected path 'pfoo', got '%s'", f.Path)
	}
	if f.EnvironmentVariable != "" {
		t.Errorf("Expected env '', got '%s'", f.EnvironmentVariable)
	}
	b := bundle.Credentials["bar"]
	if b.Path != "" {
		t.Errorf("Expected path '', got '%s'", b.Path)
	}
	if b.EnvironmentVariable != "ebar" {
		t.Errorf("Expected env 'ebar', got '%s'", b.EnvironmentVariable)
	}
	q := bundle.Credentials["quux"]
	if q.Path != "pquux" {
		t.Errorf("Expected path 'pquux', got '%s'", q.Path)
	}
	if q.EnvironmentVariable != "equux" {
		t.Errorf("Expected env 'equux', got '%s'", q.EnvironmentVariable)
	}
}
