package replacement

import (
	yaml "gopkg.in/yaml.v2"
)

// NewYAMLReplacer creates a Replacer for YAML documents.
func NewYAMLReplacer() Replacer {
	return yamlReplacer{}
}

type yamlReplacer struct {
}

func (r yamlReplacer) Replace(source string, selector string, value string) (string, error) {
	dict := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(source), dict)

	if err != nil {
		return "", err
	}

	selectorPath := parseSelector(selector)
	replaceInObjectMap(dict, selectorPath, value)

	bytes, err := yaml.Marshal(dict)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
