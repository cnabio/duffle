package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/duffle/pkg/duffle/home"
)

func setupTempDuffleHome(t *testing.T) (string, error) {
	tempDuffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		return "", err
	}
	duffleHome := home.Home(tempDuffleHome)
	if err := os.MkdirAll(duffleHome.Bundles(), 0755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(duffleHome.Logs(), 0755); err != nil {
		return "", err
	}
	keyring := filepath.Join("..", "..", "pkg", "signature", "testdata", "keyring.gpg")

	err = copyFile(keyring, filepath.Join(tempDuffleHome, "public.ring"))
	if err != nil {
		return "", err
	}
	mockSigningKeyring(tempDuffleHome, t)

	return tempDuffleHome, nil
}

func TestExportSetup(t *testing.T) {
	out := ioutil.Discard
	tempDuffleHome, err := setupTempDuffleHome(t)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempDuffleHome)

	if err := copySignedTestBundle(tempDuffleHome); err != nil {
		t.Fatal(err)
	}

	// Setup a temporary dir for destination
	tempDir, err := ioutil.TempDir("", "duffledest")
	if err != nil {
		t.Fatal(err)
	}

	duffleHome := home.Home(tempDuffleHome)
	exp := &exportCmd{
		bundle: "foo:1.0.0",
		dest:   tempDir,
		home:   duffleHome,
		out:    out,
	}

	source, _, err := exp.setup()
	if err != nil {
		t.Errorf("Did not expect error but got %s", err)
	}

	expectedSource := filepath.Join(tempDuffleHome, "bundles", "foo-1.0.0.cnab")
	if source != expectedSource {
		t.Errorf("Expected source to be %s, got %s", expectedSource, source)
	}

	expFail := &exportCmd{
		bundle: "bar:1.0.0",
		dest:   tempDir,
		home:   duffleHome,
		out:    out,
	}
	_, _, err = expFail.setup()
	if err == nil {
		t.Error("Expected error, got none")
	}

	bundlepath := filepath.Join("..", "..", "tests", "testdata", "bundles", "foo.json")
	expFile := &exportCmd{
		bundle:       bundlepath,
		dest:         tempDir,
		home:         duffleHome,
		out:          out,
		bundleIsFile: true,
	}
	source, _, err = expFile.setup()
	if err != nil {
		t.Errorf("Did not expect error but got %s", err)
	}

	if source != bundlepath {
		t.Errorf("Expected bundle file path to be %s, got %s", bundlepath, source)
	}
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return nil
	}
	return nil
}

func copySignedTestBundle(tempDuffleHome string) error {
	signedBundle := filepath.Join("..", "..", "tests", "testdata", "bundles", "foo.cnab")
	outfile := "foo-1.0.0.cnab"
	if err := copyFile(signedBundle, filepath.Join(tempDuffleHome, "bundles", outfile)); err != nil {
		return err
	}
	var jsonBlob = []byte(`{
    "foo": {
        "1.0.0": "foo-1.0.0.cnab"
        }
    } `)
	return ioutil.WriteFile(filepath.Join(tempDuffleHome, "repositories.json"), jsonBlob, 0644)
}
