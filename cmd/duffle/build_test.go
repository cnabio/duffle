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

	"github.com/scothis/ruffle/pkg/duffle/home"
	"github.com/scothis/ruffle/pkg/repo"
	"github.com/scothis/ruffle/pkg/signature"
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

	// Create temporary signing key
	mockSigningKeyring(testHome.String(), t)

	if err := cmd.run(); err != nil {
		t.Errorf("Expected no error but got err: %s", err)
	}

	// Verify that the bundle exists and is signed
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
	data, err := ioutil.ReadFile(loc)
	is.NoError(err)
	is.Contains(string(data), "---BEGIN PGP SIGNED MESSAGE----")
}

func mockSigningKeyring(tempHome string, t *testing.T) {
	t.Helper()
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
