package replacement

import (
	"errors"
	"fmt"
	"strings"

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
	replaceIn(dict, selectorPath, value)

	bytes, err := yaml.Marshal(dict)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func parseSelector(selector string) []string {
	return strings.Split(selector, ".")
}

func replaceIn(dict map[interface{}]interface{}, selectorPath []string, value string) error {
	entry, ok := dict[selectorPath[0]]
	if !ok {
		return errors.New("Selector not found")
	}

	if len(selectorPath) == 1 {
		dict[selectorPath[0]] = value
		return nil
	}

	entryDict, ok := entry.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("Entry %s is not a map", selectorPath[0])
	}
	rest := selectorPath[1:]

	return replaceIn(entryDict, rest, value)
}
