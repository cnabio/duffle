package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deislabs/duffle/pkg/repo"
)

func TestBuild(t *testing.T) {
	testHome := CreateTestHome(t)
	defer os.RemoveAll(testHome.String())

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
		home: testHome,
		src:  testBundlePath,
		out:  out,
	}

	if err := cmd.run(); err != nil {
		t.Errorf("Expected no error but got err: %s", err)
	}

	// Verify that the bundle exists
	is := assert.New(t)

	index, err := repo.LoadIndex(testHome.Repositories())
	if err != nil {
		t.Errorf("cannot open %s: %v", testHome.Repositories(), err)
	}

	// since we've only built one bundle, let's just fetch the latest version
	digest, err := index.Get("testbundle", "")
	if err != nil {
		t.Fatalf("could not find bundle: %v", err)
	}

	loc := filepath.Join(testHome.Bundles(), digest)
	is.FileExists(loc)
	_, err = ioutil.ReadFile(loc)
	is.NoError(err)
}
