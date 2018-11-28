package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/duffle/pkg/duffle/manifest"
)

func TestCreateCmd(t *testing.T) {
	name := "test-bundle"
	tdir, err := ioutil.TempDir("", "duffle-create")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	cmd := newCreateCmd(ioutil.Discard)
	path := filepath.Join(tdir, name)
	if err := cmd.RunE(cmd, []string{path}); err != nil {
		t.Fatalf("Failed to run create: %s", err)
	}

	if fi, err := os.Stat(path); err != nil {
		t.Fatalf("Nothing created at path: %s", err)
	} else if !fi.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}

	m, err := manifest.Load("duffle.json", path)
	if err != nil {
		t.Errorf("Unable to load duffle.json file: %s", err)
	}

	if m.Name != name {
		t.Errorf("Expected name of bundle to be %s, got %s", name, m.Name)
	}
}
