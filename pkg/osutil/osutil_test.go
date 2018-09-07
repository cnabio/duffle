package osutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestExists(t *testing.T) {
	file, err := ioutil.TempFile("", "osutil")
	if err != nil {
		t.Fatal(err)
	}
	name := file.Name()

	exists, err := Exists(name)
	if err != nil {
		t.Errorf("expected no error when calling Exists() on a file that exists, got %v", err)
	}
	if !exists {
		t.Error("expected tempfile to exist")
	}
	// on Windows, we need to close all open handles to a file before we remove it.
	file.Close()
	os.Remove(name)
	stillExists, err := Exists(name)
	if err != nil {
		t.Errorf("expected no error when calling Exists() on a file that does not exist, got %v", err)
	}
	if stillExists {
		t.Error("expected tempfile to NOT exist after removing it")
	}
}

func TestEnsureDirectory(t *testing.T) {
	dir, err := ioutil.TempDir("", "osutil")
	if err != nil {
		t.Fatal(err)
	}
	// cleanup after the test
	defer os.Remove(dir)
	os.Remove(dir)

	if err := EnsureDirectory(dir); err != nil {
		t.Errorf("expected no error when calling EnsureDirectory() on a directory that doesn't exist, got %v", err)
	}

	exists, err := Exists(dir)
	if err != nil {
		t.Errorf("expected no error when calling Exists() on a directory that exists, got %v", err)
	}
	if !exists {
		t.Error("expected directory to exist")
	}
}

func TestEnsureDirectoryOnFile(t *testing.T) {
	file, err := ioutil.TempFile("", "osutil")
	if err != nil {
		t.Fatal(err)
	}
	name := file.Name()
	// cleanup after the test
	defer os.Remove(name)
	file.Close()

	if err := EnsureDirectory(name); err.Error() != fmt.Sprintf("%s must be a directory", name) {
		t.Errorf("expected an error when calling EnsureDirectory() on a file, got %v", err)
	}
}

func TestEnsureFile(t *testing.T) {
	file, err := ioutil.TempFile("", "osutil")
	if err != nil {
		t.Fatal(err)
	}
	name := file.Name()
	// cleanup after the test
	defer os.Remove(name)
	file.Close()
	os.Remove(name)

	if err := EnsureFile(name); err != nil {
		t.Errorf("expected no error when calling EnsureFile() on a file that doesn't exist, got %v", err)
	}

	exists, err := Exists(name)
	if err != nil {
		t.Errorf("expected no error when calling Exists() on a file that exists, got %v", err)
	}
	if !exists {
		t.Error("expected file to exist")
	}
}

func TestEnsureFileOnDirectory(t *testing.T) {
	dir, err := ioutil.TempDir("", "osutil")
	if err != nil {
		t.Fatal(err)
	}
	// cleanup after the test
	defer os.Remove(dir)

	if err := EnsureFile(dir); err.Error() != fmt.Sprintf("%s must not be a directory", dir) {
		t.Errorf("expected an error when calling EnsureFile() on a directory that exists, got %v", err)
	}
}
