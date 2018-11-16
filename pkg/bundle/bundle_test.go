package bundle

import (
	"errors"
	"fmt"
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
	if image1.Description != "image1" {
		t.Errorf("Expected description 'image1', got '%s'", image1.Description)
	}
	if image1.Image != "urn:image1uri" {
		t.Errorf("Expected Image 'urn:image1uri', got '%s'", image1.Image)
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

func TestValuesOrDefaults_Required(t *testing.T) {
	is := assert.New(t)
	vals := map[string]interface{}{
		"enabled": true,
	}
	b := &Bundle{
		Parameters: map[string]ParameterDefinition{
			"minimum": {
				DataType: "int",
				Required: true,
			},
			"enabled": {
				DataType:     "bool",
				DefaultValue: false,
			},
		},
	}

	_, err := ValuesOrDefaults(vals, b)
	is.Error(err)

	// It is unclear what the outcome should be when the user supplies
	// empty values on purpose. For now, we will assume those meet the
	// minimum definition of "required", and that other rules will
	// correct for empty values.
	//
	// Example: It makes perfect sense for a user to specify --set minimum=0
	// and in so doing meet the requirement that a value be specified.
	vals["minimum"] = 0
	res, err := ValuesOrDefaults(vals, b)
	is.NoError(err)
	is.Equal(0, res["minimum"])
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

type fakeContainerImageResolver struct {
	resolveFun func(string, string) (string, string, error)
}

func (f fakeContainerImageResolver) Resolve(originalImageRef, originalDigest string) (string, string, error) {
	return f.resolveFun(originalImageRef, originalDigest)
}

func makeInvocationImageBundle(image, imageType, digest string) *Bundle {
	return &Bundle{
		InvocationImages: []InvocationImage{
			// Digested Docker image
			{
				BaseImage: BaseImage{
					Digest:    digest,
					Image:     image,
					ImageType: imageType,
				},
			},
		},
	}
}

func TestFixupContainerImages(t *testing.T) {
	const digest = "sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341"
	testCases := []struct {
		name     string
		bundle   *Bundle
		expected *Bundle
		resolver func(string, string) (string, string, error)
		err      error
	}{
		{
			name:     "Digested Docker invocation image",
			bundle:   makeInvocationImageBundle("my-image:mytag", "docker", digest),
			expected: makeInvocationImageBundle("my-image:mytag@"+digest, "docker", digest),
			resolver: func(image, digest string) (string, string, error) {
				return fmt.Sprintf("%s@%s", image, digest), digest, nil
			},
		},
		{
			name:     "Digested OCI invocation image",
			bundle:   makeInvocationImageBundle("my-image:mytag", "oci", digest),
			expected: makeInvocationImageBundle("my-image:mytag@"+digest, "oci", digest),
			resolver: func(image, digest string) (string, string, error) {
				return fmt.Sprintf("%s@%s", image, digest), digest, nil
			},
		},
		{
			name:     "Custom invocation image",
			bundle:   makeInvocationImageBundle("my-vm-image", "custom-vm", digest),
			expected: makeInvocationImageBundle("my-vm-image", "custom-vm", digest),
		},
		{
			name:     "Not digested Docker invocation image",
			bundle:   makeInvocationImageBundle("my-image:mytag", "docker", ""),
			expected: makeInvocationImageBundle("my-image:mytag@"+digest, "docker", digest),
			resolver: func(image, _ string) (string, string, error) { return fmt.Sprintf("%s@%s", image, digest), digest, nil },
		},
		{
			name:     "Not digested OCI invocation image",
			bundle:   makeInvocationImageBundle("my-image:mytag", "oci", ""),
			expected: makeInvocationImageBundle("my-image:mytag@"+digest, "oci", digest),
			resolver: func(image, _ string) (string, string, error) { return fmt.Sprintf("%s@%s", image, digest), digest, nil },
		},
		{
			name:     "Unresolved image",
			bundle:   makeInvocationImageBundle("my-image:mytag", "docker", ""),
			resolver: func(_, _ string) (string, string, error) { return "", "", errors.New("unresolved image") },
			err:      errors.New("unresolved image"),
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			is := assert.New(t)
			err := c.bundle.FixupContainerImages(fakeContainerImageResolver{resolveFun: c.resolver})
			if c.err != nil {
				is.EqualError(err, c.err.Error())
			} else {
				is.Equal(c.expected, c.bundle)
			}
		})
	}
}
