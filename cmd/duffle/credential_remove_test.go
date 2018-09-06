package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deis/duffle/pkg/duffle/home"
)

func TestCredentialRemove(t *testing.T) {
	duffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(duffleHome)
	credentialsDir := filepath.Join(duffleHome, "credentials")
	if err := os.MkdirAll(filepath.Join(duffleHome, "credentials"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := setupCredentialsDir(credentialsDir); err != nil {
		t.Fatal(err)
	}
	out := bytes.NewBuffer(nil)
	cmd := &credentialRemoveCmd{
		home:  home.Home(duffleHome),
		out:   out,
		names: []string{"testing", "example"},
	}

	if err := cmd.run(); err != nil {
		t.Errorf("Expected no error, got error: %s", err)
	}

	nonexistPath := filepath.Join(cmd.home.Credentials(), "testing.yaml")
	if _, err := os.Stat(nonexistPath); !os.IsNotExist(err) {
		t.Errorf("Expected %s to not to exist but does", nonexistPath)
	}

	nonexistPath = filepath.Join(cmd.home.Credentials(), "example.yaml")
	if _, err := os.Stat(nonexistPath); !os.IsNotExist(err) {
		t.Errorf("Expected %s to not to exist but does", nonexistPath)
	}

	anotherPath := filepath.Join(cmd.home.Credentials(), "another.yaml")
	if _, err := os.Stat(anotherPath); err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but does not", anotherPath)
	}

	cmd.names = []string{"another"}
	if err := cmd.run(); err != nil {
		t.Errorf("Expected no error, got error: %s", err)
	}

	if _, err := os.Stat(anotherPath); !os.IsNotExist(err) {
		t.Errorf("Expected %s to not to exist but does", anotherPath)
	}

	if err := cmd.run(); err == nil {
		t.Error("Expected error, got none")
	}
}

func setupCredentialsDir(credentialsDir string) error {
	dest := filepath.Join(credentialsDir, "testing.yaml")
	if err := copyCredentialSetFile(dest, "testdata/dufflehome/credentials/testing.yaml"); err != nil {
		return err
	}

	dest = filepath.Join(credentialsDir, "example.yaml")
	if err := copyCredentialSetFile(dest, "testdata/dufflehome/credentials/example.yaml"); err != nil {
		return err
	}

	dest = filepath.Join(credentialsDir, "another.yaml")
	if err := copyCredentialSetFile(dest, "testdata/dufflehome/credentials/another.yaml"); err != nil {
		return err
	}
	return nil
}
