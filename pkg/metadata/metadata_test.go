package metadata

import (
	"testing"
)

func TestReadTopLevelProperties(t *testing.T) {
	json := `{
		"name": "foo",
		"version": "1.0",
		"images": []
	}`
	metadata, err := Parse(json)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Name != "foo" {
		t.Errorf("Expected name 'foo', got '%s'", metadata.Name)
	}
	if metadata.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", metadata.Version)
	}
	if len(metadata.Images) != 0 {
		t.Errorf("Expected no images, got %d", len(metadata.Images))
	}
}

func TestReadImageProperties(t *testing.T) {
	json := `{
		"name": "foo",
		"version": "1.0",
		"images": [
			{
				"name": "image1",
				"uri": "urn:image1uri",
				"refs": [
					{
						"path": "image1path",
						"field": "image.1.field"
					}
				]
			},
			{
				"name": "image2",
				"uri": "urn:image2uri",
				"refs": [
					{
						"path": "image2path",
						"field": "image.2.field"
					}
				]
			}
		]
	}`
	metadata, err := Parse(json)
	if err != nil {
		t.Fatal(err)
	}
	if len(metadata.Images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(metadata.Images))
	}
	image1 := metadata.Images[0]
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
