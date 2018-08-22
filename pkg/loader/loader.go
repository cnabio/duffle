package loader

import (
	"net/url"
	"os"

	"github.com/deis/duffle/pkg/bundle"
)

// Loader provides an interface for loading a bundle
type Loader interface {
	Load() (bundle.Bundle, error)
}

// New determines and returns the correct Loader for the given bundle file
func New(bundleFile string) (Loader, error) {
	if isLocalReference(bundleFile) {
		return LocalLoader{source: bundleFile}, nil
	}

	_, err := url.ParseRequestURI(bundleFile)
	if err != nil {
		return nil, err
	}
	return RemoteLoader{source: bundleFile}, nil
}

func isLocalReference(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}
