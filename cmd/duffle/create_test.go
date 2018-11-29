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

	// also test the output of duffle.json, since manifest.Load won't necessarily catch that
	mbytes, err := ioutil.ReadFile(filepath.Join(path, "duffle.json"))
	if err != nil {
		t.Error(err)
	}

	expected := `{
    "name": "test-bundle",
    "version": "0.1.0",
    "description": "A short description of your bundle",
    "keywords": [
        "test-bundle",
        "cnab",
        "tutorial"
    ],
    "maintainers": [
        {
            "name": "John Doe",
            "email": "john.doe@example.com",
            "url": "https://example.com"
        },
        {
            "name": "Jane Doe",
            "email": "jane.doe@example.com",
            "url": "https://example.com"
        }
    ],
    "components": {
        "cnab": {
            "name": "cnab",
            "builder": "docker",
            "configuration": {
                "registry": "microsoft"
            }
        }
    },
    "parameters": null,
    "credentials": null
}`
	if string(mbytes) != expected {
		t.Errorf("Expected duffle.json output to look like this:\n\n%s\n\nGot:\n\n%s", expected, string(mbytes))
	}
}
