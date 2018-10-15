package loader

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// remoteFetcher downloads a remote file containing a bundle and returns it
type remoteFetcher string

// Load retrieves a remote CNAB bundle and deserializes it
func (l remoteFetcher) Bytes() ([]byte, error) {
	response, err := http.Get(string(l))
	if err != nil {
		return []byte{}, fmt.Errorf("cannot download bundle file: %v", err)
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
