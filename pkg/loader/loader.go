package loader

import (
	"os"

	"github.com/deis/duffle/pkg/bundle"
)

// Loader provides an interface for loading a bundle
type Loader interface {
	Load() (bundle.Bundle, error)
}

// New determines and returns the correct Loader for the given bundle file
func New(bundleFile string) (Loader, error) {
	var l Loader
	if isLocalReference(bundleFile) {
		l = LocalLoader{source: bundleFile}
	}

	return l, nil
}

func isLocalReference(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}
