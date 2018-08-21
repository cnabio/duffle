package replacement

import (
	"errors"
	"fmt"
	"strings"
)

func parseSelector(selector string) []string {
	return strings.Split(selector, ".")
}

func replaceInObjectMap(dict map[interface{}]interface{}, selectorPath []string, value string) error {
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

	return replaceInObjectMap(entryDict, rest, value)
}

// TODO: This duplication makes me VERY ANGRY
func replaceInStringMap(dict map[string]interface{}, selectorPath []string, value string) error {
	entry, ok := dict[selectorPath[0]]
	if !ok {
		return errors.New("Selector not found")
	}

	if len(selectorPath) == 1 {
		dict[selectorPath[0]] = value
		return nil
	}

	entryDict, ok := entry.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Entry %s is not a map", selectorPath[0])
	}
	rest := selectorPath[1:]

	return replaceInStringMap(entryDict, rest, value)
}
