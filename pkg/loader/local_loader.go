package loader

import (
	"fmt"
	"io/ioutil"

	"github.com/deis/duffle/pkg/bundle"
)

// LocalLoader loads a bundle form a local file
type LocalLoader struct {
	source string
}

// Load reads a local CNAB bundle and deserializes it
func (l LocalLoader) Load() (bundle.Bundle, error) {
	var b bundle.Bundle

	data, err := ioutil.ReadFile(l.source)
	if err != nil {
		return b, fmt.Errorf("cannot read bundle file: %v", err)
	}

	return bundle.ParseBuffer(data)
}
