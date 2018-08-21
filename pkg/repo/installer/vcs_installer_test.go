package installer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/repo"
)

var _ Installer = new(VCSInstaller)

func TestVCSInstallerSuccess(t *testing.T) {
	// TODO: run this against a mock vcs repo
	t.Skip("duffle-bundles doesn't exist")
	dh, err := ioutil.TempDir("", "fish-home-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dh)

	home := home.Home(dh)
	if err := os.MkdirAll(home.Repositories(), 0755); err != nil {
		t.Fatalf("Could not create %s: %s", home.Repositories(), err)
	}

	source := "https://github.com/deis/duffle-bundles"
	i, err := New(source, "", "", home)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	expectedPath := home.Path("repositories", "github.com", "deis", "duffle-bundles")
	if i.Path() != expectedPath {
		t.Errorf("expected path '%s', got %q", expectedPath, i.Path())
	}

	// ensure a VCSInstaller was returned
	vi, ok := i.(*VCSInstaller)
	if !ok {
		t.Error("expected a VCSInstaller")
	}

	expectedName := "github.com/deis/duffle-bundles"
	if vi.Name != expectedName {
		t.Errorf("expected name '%s', got '%s'", expectedName, vi.Name)
	}

	vi.Name = "foo"

	expectedPath = home.Path("repositories", "foo")
	if i.Path() != expectedPath {
		t.Errorf("expected path '%s', got %q", expectedPath, i.Path())
	}

	if err := Install(i); err != nil {
		t.Error(err)
	}
}

func TestVCSInstallerUpdate(t *testing.T) {
	// TODO: run this against a mock vcs repo
	t.Skip("duffle-bundles doesn't exist")
	dh, err := ioutil.TempDir("", "vcsinstaller-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dh)

	home := home.Home(dh)
	if err := os.MkdirAll(home.Repositories(), 0755); err != nil {
		t.Fatalf("Could not create %s: %s", home.Repositories(), err)
	}

	source := "https://github.com/deis/duffle-bundles"
	i, err := New(source, "", "", home)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// ensure a VCSInstaller was returned
	_, ok := i.(*VCSInstaller)
	if !ok {
		t.Error("expected a VCSInstaller")
	}

	if err := Update(i); err == nil {
		t.Error("expected error for repository does not exist, got none")
	} else if err.Error() != "repository does not exist" {
		t.Errorf("expected error for repository does not exist, got (%v)", err)
	}

	// Install repository before update
	if err := Install(i); err != nil {
		t.Error(err)
	}

	// Update repository
	if err := Update(i); err != nil {
		t.Error(err)
	}

	// Test update failure
	os.Remove(filepath.Join(i.Path(), "LICENSE"))
	// Testing update for error
	if err := Update(i); err == nil {
		t.Error("expected error for repository modified, got none")
	} else if err != repo.ErrRepoDirty {
		t.Errorf("expected error for repository modified, got (%v)", err)
	}

}
