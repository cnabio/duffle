package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundleSign(t *testing.T) {
	tmp, err := ioutil.TempFile("", "duffle-")
	if err != nil {
		t.Fatal(err)
	}
	outfile := tmp.Name()
	defer os.Remove(outfile)

	bundlejson := filepath.Join("..", "..", "tests", "testdata", "bundles", "foo.json")
	keyring := filepath.Join("..", "..", "pkg", "signature", "testdata", "keyring.gpg")
	identity := "test2@example.com"

	if err := signFile(bundlejson, keyring, identity, outfile, false); err != nil {
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
