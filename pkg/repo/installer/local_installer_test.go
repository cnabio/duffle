package installer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deis/duffle/pkg/duffle/home"
)

var _ Installer = new(LocalInstaller)

func TestLocalInstaller(t *testing.T) {
	dh, err := ioutil.TempDir("", "localinstaller-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dh)

	home := home.Home(dh)
	if err := os.MkdirAll(home.Repositories(), 0755); err != nil {
		t.Fatalf("Could not create %s: %s", home.Repositories(), err)
	}

	source := filepath.Join("..", "..", "..", "tests", "testdata", "home", "repositories", home.DefaultRepository())
	i, err := New(source, "", "", home)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := Install(i); err != nil {
		t.Error(err)
	}

	expectedPath := home.Path("repositories", "duffle-bundles")
	if i.Path() != expectedPath {
		t.Errorf("expected path '%s', got %q", expectedPath, i.Path())
	}
}
