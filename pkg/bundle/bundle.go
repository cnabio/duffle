package bundle

import (
	"encoding/json"
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
	Name            string                         `json:"name" toml:"name"`
	Version         string                         `json:"version" toml:"version"`
	InvocationImage InvocationImage                `json:"invocationImage" toml:"invocationImage"`
	Images          []Image                        `json:"images" toml:"images"`
	Parameters      map[string]ParameterDefinition `json:"parameters" toml:"parameters"`
	Credentials     map[string]CredentialLocation  `json:"credentials" toml:"credentials"`
}
