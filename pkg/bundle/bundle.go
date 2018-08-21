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
	Path  string `json:"path"`
	Field string `json:"field"`
}

// Image describes a container image in the bundle
type Image struct {
	Name string        `json:"name"`
	URI  string        `json:"uri"`
	Refs []LocationRef `json:"refs"`
}

// Bundle is a CNAB metadata document
type Bundle struct {
	Name       string                         `json:"name"`
	Version    string                         `json:"version"`
	Images     []Image                        `json:"images"`
	Parameters map[string]ParameterDefinition `json:"parameters"`
}
