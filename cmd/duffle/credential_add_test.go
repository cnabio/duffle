package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deis/duffle/pkg/duffle/home"
)

func TestCredentialAddFile(t *testing.T) {
	duffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(duffleHome)
	if err := os.MkdirAll(filepath.Join(duffleHome, "credentials"), 0755); err != nil {
		t.Fatal(err)
	}

	out := bytes.NewBuffer(nil)
	cmd := &credentialAddCmd{
		home:  home.Home(duffleHome),
		out:   out,
		paths: []string{"testdata/dufflehome/credentials/testing.yaml"},
	}

	if err := cmd.run(); err != nil {
		t.Fatal(err)
	}

	expectedPath := filepath.Join(cmd.home.Credentials(), "testing.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to be created but was not", expectedPath)
	}

}

func TestCredentialAddMultipleFiles(t *testing.T) {
	duffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(duffleHome)
	if err := os.MkdirAll(filepath.Join(duffleHome, "credentials"), 0755); err != nil {
		t.Fatal(err)
	}

	out := bytes.NewBuffer(nil)
	cmd := &credentialAddCmd{
		home:  home.Home(duffleHome),
		out:   out,
		paths: []string{"testdata/dufflehome/credentials/testing.yaml", "testdata/dufflehome/credentials/example.yaml"},
	}

	if err := cmd.run(); err != nil {
		t.Fatal(err)
	}

	expectedPath := filepath.Join(cmd.home.Credentials(), "testing.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to be created but was not", expectedPath)
	}

	expectedPath = filepath.Join(cmd.home.Credentials(), "example.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected %s to be created but was not", expectedPath)
	}

}

func TestCredentialAddMalformedFile(t *testing.T) {
	duffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(duffleHome)
	if err := os.MkdirAll(filepath.Join(duffleHome, "credentials"), 0755); err != nil {
		t.Fatal(err)
	}

	out := bytes.NewBuffer(nil)
	cmd := &credentialAddCmd{
		home:  home.Home(duffleHome),
		out:   out,
		paths: []string{"testdata/malformedhome/credentials/malformed.yaml"},
	}

	err = cmd.run()
	if err == nil {
		t.Error("Expected error but got none")
	}

	expected := fmt.Sprintf("%s is not a valid credential set", cmd.paths[0])
	if fmt.Sprintf("%s", err) != expected {
		t.Errorf("Expected err `%s`, got `%s`", expected, err)
	}

	nonexistPath := filepath.Join(cmd.home.Credentials(), "malformed.yaml")
	if _, err := os.Stat(nonexistPath); !os.IsNotExist(err) {
		t.Errorf("Expected %s to not exist", nonexistPath)
	}
}

func TestCredentialAddInvalidFileName(t *testing.T) {
	duffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(duffleHome)
	if err := os.MkdirAll(filepath.Join(duffleHome, "credentials"), 0755); err != nil {
		t.Fatal(err)
	}

	out := bytes.NewBuffer(nil)
	cmd := &credentialAddCmd{
		home:  home.Home(duffleHome),
		out:   out,
		paths: []string{"testdata/malformedhome/credentials/invalid.yaml"},
	}

	err = cmd.run()
	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestCredentialAddDuplicate(t *testing.T) {
	//setup
	duffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(duffleHome)
	credentialsDir := filepath.Join(duffleHome, "credentials")
	if err := os.MkdirAll(filepath.Join(duffleHome, "credentials"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Create(filepath.Join(credentialsDir, "testing.yaml")); err != nil {
		t.Fatal(err)
	}

	//create cmd
	out := bytes.NewBuffer(nil)
	cmd := &credentialAddCmd{
		home:  home.Home(duffleHome),
		out:   out,
		paths: []string{"testdata/dufflehome/credentials/testing.yaml"},
	}

	//run cmd
	if err := cmd.run(); err == nil {
		t.Error("Expected duplication error, got no error")
	}
}
