package loader

import (
	"fmt"
	"net/url"
	"os"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/signature"
)

// Loader provides an interface for loading a bundle
type Loader interface {
	Load(source string) (*bundle.Bundle, error)
}

// New creates a loader for signed bundle files.
func New(keyring *signature.KeyRing) Loader {
	return &SecureLoader{
		keyring: keyring,
	}
}

// NewInsecure loads an unsigned bundle.json
func NewInsecure() Loader {
	return &InsecureLoader{}
}

type Fetcher interface {
	Bytes() ([]byte, error)
}

func getFetcher(bundleFile string) (Fetcher, error) {
	if isLocalReference(bundleFile) {
		return localFetcher(bundleFile), nil
	}

	if _, err := url.ParseRequestURI(bundleFile); err != nil {
		// The error emitted by ParseRequestURI is icky.
		return nil, fmt.Errorf("bundle %q not found", bundleFile)
	}
	return remoteFetcher(bundleFile)
}

func isLocalReference(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}
