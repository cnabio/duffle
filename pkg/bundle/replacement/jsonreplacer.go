package replacement

import (
	"encoding/json"
)

// NewJSONReplacer creates a Replacer for JSON documents.
func NewJSONReplacer(indent string) Replacer {
	return jsonReplacer{
		indent: indent,
	}
}

type jsonReplacer struct {
	indent string
}

func (r jsonReplacer) Replace(source string, selector string, value string) (string, error) {
	dict := make(map[string]interface{})
	err := json.Unmarshal([]byte(source), &dict)

	if err != nil {
		return "", err
	}

	selectorPath := parseSelector(selector)
	replaceInStringMap(dict, selectorPath, value)

	bytes, err := json.MarshalIndent(dict, "", r.indent)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
