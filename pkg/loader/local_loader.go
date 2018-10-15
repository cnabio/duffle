package loader

import (
	"io/ioutil"
)

// localLoader loads a bundle form a local file
type localFetcher string

// Load reads a local CNAB bundle and deserializes it
func (l localFetcher) Bytes() ([]byte, error) {
	return ioutil.ReadFile(string(l))
}
