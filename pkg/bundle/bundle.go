package bundle

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ParseBuffer reads CNAB metadata out of a JSON byte stream
func ParseBuffer(data []byte) (Bundle, error) {
	b := Bundle{}
	err := json.Unmarshal(data, &b)
	return b, err
}

// Parse reads CNAB metadata from a JSON string
func Parse(text string) (Bundle, error) {
	return ParseBuffer([]byte(text))
}

// LocationRef specifies a location within the invocation package
type LocationRef struct {
	Path  string `json:"path" toml:"path"`
	Field string `json:"field" toml:"field"`
}

// Image describes a container image in the bundle
type Image struct {
	Name string        `json:"name" toml:"name"`
	URI  string        `json:"uri" toml:"uri"`
	Refs []LocationRef `json:"refs" toml:"refs"`
}

// InvocationImage contains the image type and location for the installation of a bundle
type InvocationImage struct {
	ImageType string `json:"imageType" toml:"imageType"`
	Image     string `json:"image" toml:"image"`
}

// CredentialLocation provides the location of a credential that the invocation
// image needs to use.
type CredentialLocation struct {
	Path                string `json:"path" toml:"path"`
	EnvironmentVariable string `json:"env" toml:"env"`
}

// Bundle is a CNAB metadata document
type Bundle struct {
	Name             string                         `json:"name" toml:"name"`
	Version          string                         `json:"version" toml:"version"`
	InvocationImages []InvocationImage              `json:"invocationImages" toml:"invocationImages"`
	Images           []Image                        `json:"images" toml:"images"`
	Parameters       map[string]ParameterDefinition `json:"parameters" toml:"parameters"`
	Credentials      map[string]CredentialLocation  `json:"credentials" toml:"credentials"`
}

// ValuesOrDefaults returns parameter values or the default parameter values
func ValuesOrDefaults(vals map[string]interface{}, b *Bundle) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	for name, def := range b.Parameters {
		if val, ok := vals[name]; ok {
			if err := def.ValidateParameterValue(val); err != nil {
				return res, fmt.Errorf("can't use %v as value of %s: %s", val, name, err)
			}
			typedVal := def.CoerceValue(val)
			res[name] = typedVal
			continue
		}
		res[name] = def.DefaultValue
	}
	return res, nil
}

// Validate the bundle contents.
func (b Bundle) Validate() error {
	if len(b.InvocationImages) == 0 {
		return errors.New("at least one invocation image must be defined in the bundle")
	}

	for _, img := range b.InvocationImages {
		err := img.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// Validate the image contents.
func (img InvocationImage) Validate() error {
	switch img.ImageType {
	case "docker", "oci":
		return validateDockerish(img.Image)
	default:
		return nil
	}
}

func validateDockerish(s string) error {
	if !strings.Contains(s, ":") {
		return errors.New("version is required")
	}
	return nil
}
