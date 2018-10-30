package bundle

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadTopLevelProperties(t *testing.T) {
	json := `{
		"name": "foo",
		"version": "1.0",
		"images": [],
		"credentials": {}
	}`
	bundle, err := Unmarshal([]byte(json))
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
	data, err := ioutil.ReadFile("../../tests/testdata/bundles/foo.json")
	if err != nil {
		t.Errorf("cannot read bundle file: %v", err)
	}

	bundle, err := Unmarshal(data)
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
	data, err := ioutil.ReadFile("../../tests/testdata/bundles/foo.json")
	if err != nil {
		t.Errorf("cannot read bundle file: %v", err)
	}

	bundle, err := Unmarshal(data)
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

func TestValuesOrDefaults(t *testing.T) {
	is := assert.New(t)
	vals := map[string]interface{}{
		"port":    8080,
		"host":    "localhost",
		"enabled": true,
	}
	b := &Bundle{
		Parameters: map[string]ParameterDefinition{
			"port": {
				DataType:     "int",
				DefaultValue: 1234,
			},
			"host": {
				DataType:     "string",
				DefaultValue: "localhost.localdomain",
			},
			"enabled": {
				DataType:     "bool",
				DefaultValue: false,
			},
			"replicaCount": {
				DataType:     "int",
				DefaultValue: 3,
			},
		},
	}

	vod, err := ValuesOrDefaults(vals, b)

	is.NoError(err)
	is.True(vod["enabled"].(bool))
	is.Equal(vod["host"].(string), "localhost")
	is.Equal(vod["port"].(int), 8080)
	is.Equal(vod["replicaCount"].(int), 3)

	// This should err out because of type problem
	vals["replicaCount"] = "banana"
	_, err = ValuesOrDefaults(vals, b)
	is.Error(err)
}

func TestValidateBundle_RequiresInvocationImage(t *testing.T) {
	b := Bundle{
		Name:    "bar",
		Version: "0.1.0",
	}

	err := b.Validate()
	if err == nil {
		t.Fatal("Validate should have failed because the bundle has no invocation images")
	}

	b.InvocationImages = append(b.InvocationImages, InvocationImage{})

	err = b.Validate()
	if err != nil {
		t.Fatal(err)
	}
}
