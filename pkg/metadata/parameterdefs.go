package metadata

import (
	"encoding/json"
	"fmt"
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
	MinValue      *int              `json:"minValue"`
	MaxValue      *int              `json:"maxValue"`
	MinLength     *int              `json:"minLength"`
	MaxLength     *int              `json:"maxLength"`
	Metadata      ParameterMetadata `json:"metadata"`
}

// ParameterMetadata contains metadata for a parameter definition.
type ParameterMetadata struct {
	Description string `json:"description"`
}

// ValidateParameterValue checks whether a value is valid as the value of
// the specified parameter.
func (pd ParameterDefinition) ValidateParameterValue(value interface{}) error {
	if len(pd.AllowedValues) > 0 {
		if !isInCollection(value, pd.AllowedValues) {
			return fmt.Errorf("Value is not in the set of allowed values for this parameter")
		}
	}

	switch pd.DataType {
	case "string":
		return pd.validateStringParameterValue(value)
	case "int":
		return pd.validateIntParameterValue(value)
	case "bool":
		return pd.validateBoolParameterValue(value)
	default:
		return fmt.Errorf("Invalid parameter definition")
	}
}

func (pd ParameterDefinition) validateStringParameterValue(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("Value is not a string")
	}
	if pd.MinLength != nil && len(s) < *pd.MinLength {
		return fmt.Errorf("Value is too short: minimum length is %d", pd.MinLength)
	}
	if pd.MaxLength != nil && len(s) > *pd.MaxLength {
		return fmt.Errorf("Value is too long: maximum length is %d", pd.MaxLength)
	}
	return nil
}

func (pd ParameterDefinition) validateIntParameterValue(value interface{}) error {
	i, ok := value.(int)
	if !ok {
		f, ok := value.(float64)
		if !ok {
			return fmt.Errorf("Value is not a number")
		}
		i, ok = asInt(f)
		if !ok {
			return fmt.Errorf("Value is not an integer")
		}
	}
	if pd.MinValue != nil && i < *pd.MinValue {
		return fmt.Errorf("Value is too low: minimum value is %d", pd.MinValue)
	}
	if pd.MaxValue != nil && i > *pd.MaxValue {
		return fmt.Errorf("Value is too long: maximum length is %d", pd.MaxValue)
	}
	return nil
}

func (pd ParameterDefinition) validateBoolParameterValue(value interface{}) error {
	_, ok := value.(bool)
	if !ok {
		return fmt.Errorf("Value is not a string")
	}
	return nil
}

func isInCollection(value interface{}, values []interface{}) bool {
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}

func asInt(f float64) (int, bool) {
	i := int(f)
	if float64(i) != f {
		return 0, false
	}
	return i, true
}
