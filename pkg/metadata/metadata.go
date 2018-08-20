package metadata

import (
	"encoding/json"
)

// ParseBuffer reads CNAB metadata out of a JSON byte stream
func ParseBuffer(data []byte) (Metadata, error) {
	metadata := Metadata{}
	err := json.Unmarshal(data, &metadata)
	return metadata, err
}

// Parse reads CNAB metadata from a JSON string
func Parse(text string) (Metadata, error) {
	return ParseBuffer([]byte(text))
}

// LocationRef specifies a location within the invocation package
type LocationRef struct {
	Path  string `json:"path"`
	Field string `json:"field"`
}

// Image describes a container image in the bundle
type Image struct {
	Name string        `json:"name"`
	URI  string        `json:"uri"`
	Refs []LocationRef `json:"refs"`
}

// Metadata is a CNAB metadata document
type Metadata struct {
	Name       string                         `json:"name"`
	Version    string                         `json:"version"`
	Images     []Image                        `json:"images"`
	Parameters map[string]ParameterDefinition `json:"parameters"`
}
