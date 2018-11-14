package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
)

var (
	// ErrNoRepoName indicates that a repository with the given name is not found.
	ErrNoRepoName = errors.New("no repository with the given name found")
)

// RepositoryIndex defines a list of repositories with each repository's login credentials.
type RepositoryIndex map[string]RepositoryCredentials

// RepositoryCredentials defines how one authenticates against a repository.
type RepositoryCredentials struct {
	Token string `json:"token"`
}

// Load takes a file at the given path and returns a RepositoryIndex object
func Load(path string) (RepositoryIndex, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return load(f)
}

// LoadReader takes a reader and returns a RepositoryIndex object
func LoadReader(r io.Reader) (RepositoryIndex, error) {
	return load(r)
}

// LoadBuffer reads repository metadata from a JSON byte stream
func LoadBuffer(data []byte) (RepositoryIndex, error) {
	return load(bytes.NewBuffer(data))
}

// Add adds a new entry to the index
func (ri *RepositoryIndex) Add(name string, creds RepositoryCredentials) {
	var i string
	url, err := url.ParseRequestURI(name)
	if err != nil {
		i = name
	} else {
		i = url.Hostname()
	}

	(*ri)[i] = creds
}

// Remove removes an entry from the index
func (ri *RepositoryIndex) Remove(name string) {
	var i string
	url, err := url.ParseRequestURI(name)
	if err != nil {
		i = name
	} else {
		i = url.Hostname()
	}

	delete(*ri, i)
}

// Has returns true if the index has an entry for a bundle with the given name and exact version.
func (ri *RepositoryIndex) Has(name string) bool {
	_, err := ri.Get(name)
	return err == nil
}

// Get returns the digest for the given name.
//
// If version is empty, this will return the digest for the bundle with the highest version.
func (ri *RepositoryIndex) Get(name string) (RepositoryCredentials, error) {
	var i string
	url, err := url.ParseRequestURI(name)
	if err != nil {
		i = name
	} else {
		i = url.Hostname()
	}

	rc, ok := (*ri)[i]
	if !ok {
		return RepositoryCredentials{}, ErrNoRepoName
	}

	return rc, nil
}

// WriteFile writes an index file to the given destination path.
//
// The mode on the file is set to 'mode'.
func (ri *RepositoryIndex) WriteFile(dest string, mode os.FileMode) error {
	b, err := json.MarshalIndent(ri, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, b, mode)
}

// load loads an index file and does minimal validity checking.
func load(r io.Reader) (RepositoryIndex, error) {
	i := RepositoryIndex{}
	if err := json.NewDecoder(r).Decode(&i); err != nil && err != io.EOF {
		return i, err
	}
	return i, nil
}
