package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/scothis/ruffle/pkg/duffle/home"
)

func TestBundleSign(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "duffle-home")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tempDir, "bundles"), 0755); err != nil {
		t.Fatal(err)
	}

	tmp, err := ioutil.TempFile("", "duffle-")
	if err != nil {
		t.Fatal(err)
	}
	outfile := tmp.Name()
	defer os.Remove(outfile)

	bundlejson := filepath.Join("..", "..", "tests", "testdata", "bundles", "foo.json")
	keyring := filepath.Join("..", "..", "pkg", "signature", "testdata", "keyring.gpg")
	identity := "test2@example.com"

	cmd := bundleSignCmd{
		outfile:  outfile,
		identity: identity,
		out:      ioutil.Discard,
		home:     home.Home(tempDir),
	}
	if err := cmd.signBundle(bundlejson, keyring); err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadFile(outfile)
	if err != nil {
		t.Fatal(err)
	}

	sig := string(data)
	is := assert.New(t)
	is.Contains(sig, "-----BEGIN PGP SIGNATURE-----")
}
