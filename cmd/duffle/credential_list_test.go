package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cnabio/duffle/pkg/duffle/home"
)

func TestCredentialList(t *testing.T) {
	duffleHome = "testdata/dufflehome"
	out := bytes.NewBuffer(nil)
	cmd := &credentialListCmd{
		home: home.Home(duffleHome),
		out:  out,
	}

	if err := cmd.run(); err != nil {
		t.Fatal(err)
	}
	result := out.String()

	if !strings.Contains(result, "testing") {
		t.Errorf("Expected list to contain %s but did not", "testing")
	}
	if !strings.Contains(result, "example") {
		t.Errorf("Expected list to contain %s but did not", "example")
	}

	if !strings.Contains(result, "another") {
		t.Errorf("Expected list to contain %s but did not", "another")
	}
}

func TestCredentialListEmpty(t *testing.T) {
	duffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(duffleHome)
	if err := os.Mkdir(filepath.Join(duffleHome, "credentials"), 0644); err != nil {
		t.Fatal(err)
	}

	out := bytes.NewBuffer(nil)
	cmd := &credentialListCmd{
		home:  home.Home(duffleHome),
		out:   out,
		short: true,
	}

	if err := cmd.run(); err != nil {
		t.Fatal(err)
	}
	result := out.String()
	if result != "" {
		t.Errorf("Expected empty result, got %v", result)
	}
}

func TestCredentialListErrors(t *testing.T) {
	duffleHome = "testdata/malformedhome"
	out := bytes.NewBuffer(nil)
	cmd := &credentialListCmd{
		home:  home.Home(duffleHome),
		out:   out,
		short: true,
	}

	if err := cmd.run(); err != nil {
		t.Fatal(err)
	}

	result := out.String()
	expected := "example\n"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
