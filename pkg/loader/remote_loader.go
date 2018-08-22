package loader

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/deis/duffle/pkg/bundle"
)

// RemoteLoader downloads a remote file containg a bundle and returns it
type RemoteLoader struct {
	source string
}

// Load retrieves a remote CNAB bundle and deserializes it
func (l RemoteLoader) Load() (bundle.Bundle, error) {
	var b bundle.Bundle

	response, err := http.Get(l.source)
	if err != nil {
		return b, fmt.Errorf("cannot download bundle file: %v", err)
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return b, fmt.Errorf("cannot read response body: %v", err)
	}

	return bundle.ParseBuffer(data)
}
