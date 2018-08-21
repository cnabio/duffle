package crud

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// NewFileSystemStore creates a Store backed by a file system directory.
// Each key is represented by a file in that directory.
func NewFileSystemStore(baseDirectory string, fileExtension string) Store {
	return fileSystemStore{
		baseDirectory: baseDirectory,
		fileExtension: fileExtension,
	}
}

type fileSystemStore struct {
	baseDirectory string
	fileExtension string
}

func (s fileSystemStore) Store(name string, data []byte) error {
	if err := s.ensure(); err != nil {
		return err
	}

	filename := s.fileNameOf(name)

	return ioutil.WriteFile(filename, data, os.ModePerm)
}

func (s fileSystemStore) Read(name string) ([]byte, error) {
	if err := s.ensure(); err != nil {
		return nil, err
	}

	filename := s.fileNameOf(name)

	return ioutil.ReadFile(filename)
}

func (s fileSystemStore) Delete(name string) error {
	if err := s.ensure(); err != nil {
		return err
	}

	filename := s.fileNameOf(name)

	return os.Remove(filename)
}

func (s fileSystemStore) fileNameOf(name string) string {
	return filepath.Join(s.baseDirectory, fmt.Sprintf("%s.%s", name, s.fileExtension))
}

func (s fileSystemStore) ensure() error {
	fi, err := os.Stat(s.baseDirectory)
	if err == nil {
		if fi.IsDir() {
			return nil
		}
		return errors.New("Storage directory name exists, but is not a directory")
	}
	return os.MkdirAll(s.baseDirectory, os.ModePerm)
}
