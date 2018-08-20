package metadata

import (
	"encoding/json"
)

// ParseParameterDefinitionsBuffer reads CNAB parameter definitions from a JSON byte stream
func ParseParameterDefinitionsBuffer(data []byte) (ParameterDefinitions, error) {
	definitions := ParameterDefinitions{}
	err := json.Unmarshal(data, &definitions)
	return definitions, err
}

// ParseParameterDefinitions reads CNAB parameter definitions from a JSON string
func ParseParameterDefinitions(text string) (ParameterDefinitions, error) {
	return ParseParameterDefinitionsBuffer([]byte(text))
}

// ParameterDefinitions describes a set of parameters defined for
// a CNAB bundle
type ParameterDefinitions struct {
	Parameters map[string]ParameterDefinition `json:"parameters"`
}

// ParameterDefinition defines a single parameter for a CNAB bundle
type ParameterDefinition struct {
	DataType      string            `json:"type"`
	DefaultValue  interface{}       `json:"defaultValue"`
	AllowedValues []interface{}     `json:"allowedValues"`
	MinValue      int               `json:"minValue"`
	MaxValue      int               `json:"maxValue"`
	MinLength     int               `json:"minLength"`
	MaxLength     int               `json:"maxLength"`
	Metadata      ParameterMetadata `json:"metadata"`
}

// ParameterMetadata contains metadata for a parameter definition.
type ParameterMetadata struct {
	Description string `json:"description"`
}
