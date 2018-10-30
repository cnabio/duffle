package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/signature"
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

	// Create temporary signing key
	mockSigningKeyring(tempHome, t)

	if err := cmd.run(); err != nil {
		t.Errorf("Expected no error but got err: %s", err)
	}

	// Verify that the bundle.cnab exists, and is signed
	is := assert.New(t)
	loc := filepath.Join(testBundlePath, "cnab", "bundle.cnab")
	is.FileExists(loc)
	data, err := ioutil.ReadFile(loc)
	is.NoError(err)
	is.Contains(string(data), "---BEGIN PGP SIGNED MESSAGE----")
}

func mockSigningKeyring(tempHome string, t *testing.T) {
	uid, err := signature.ParseUserID("fake <fake@example.com>")
	if err != nil {
		t.Fatal(err)
	}
	ring := signature.CreateKeyRing(func(a string) ([]byte, error) { return nil, errors.New("not implemented") })
	key, err := signature.CreateKey(uid)
	if err != nil {
		t.Fatal(err)
	}
	ring.AddKey(key)
	ring.SavePrivate(home.Home(tempHome).SecretKeyRing(), true)
}
