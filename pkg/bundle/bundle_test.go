package bundle

import (
	"io/ioutil"
	"testing"
)

func TestReadTopLevelProperties(t *testing.T) {
	json := `{
		"name": "foo",
		"version": "1.0",
		"images": []
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
}

func TestReadImageProperties(t *testing.T) {
	data, err := ioutil.ReadFile("../../tests/testdata/home/repositories/github.com/deis/duffle-bundles/bundles/foo.json")
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
