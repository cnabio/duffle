package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deis/duffle/pkg/duffle/home"
)

func TestBuild(t *testing.T) {
	tempHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(duffleHome)

	tempBundleDir, err := ioutil.TempDir("", "dufflebundles")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempBundleDir)
	testBundlePath := filepath.Join(tempBundleDir, "testbundle")
	if err := os.MkdirAll(filepath.Join(testBundlePath, "cnab"), 0755); err != nil {
		t.Fatal(err)
	}
	from, err := os.Open(filepath.Join("testdata", "testbundle", "duffle.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer from.Close()
	dest := filepath.Join(testBundlePath, "duffle.json")
	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer to.Close()
	_, err = io.Copy(to, from)
	if err != nil {
		t.Fatal(err)
	}
	from.Close()
	to.Close()

	out := bytes.NewBuffer(nil)
	cmd := &buildCmd{
		home: home.Home(tempHome),
		src:  testBundlePath,
		out:  out,
	}

	if err := cmd.run(); err != nil {
		t.Errorf("Expected no error but got err: %s", err)
	}
}
