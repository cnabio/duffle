package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/duffle/pkg/duffle/home"
)

func TestBundleRemove(t *testing.T) {
	tempDuffleHome, err := ioutil.TempDir("", "duffle-home")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempDuffleHome)
	duffleHome := home.Home(tempDuffleHome)
	if err := os.MkdirAll(duffleHome.Bundles(), 0755); err != nil {
		t.Fatal(err)
	}
	if err := copySignedTestBundle(tempDuffleHome); err != nil {
		t.Fatal(err)
	}

	out := ioutil.Discard
	cmd := bundleRemoveCmd{
		home:      duffleHome,
		bundleRef: "foo",
		out:       out,
	}
	if err := cmd.run(); err != nil {
		t.Errorf("Did not expect error, got %s", err)
	}

	if _, err := os.Stat(filepath.Join(cmd.home.Bundles(), "foo-1.0.0.cnab")); !os.IsNotExist(err) {
		t.Errorf("Expected bundle file to be removed from local store but was not")
	}
}
