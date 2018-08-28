package installer

import (
	"os"
	"path/filepath"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/repo"
)

// Installer provides an interface for installing client repositories.
type Installer interface {
	// Install adds a repository to a path
	Install() error
	// Path is the directory of the installed repo.
	Path() string
	// Update updates a repo.
	Update() error
}

// Install installs a repo.
func Install(i Installer) error {
	basePath := filepath.Dir(i.Path())
	if _, pathErr := os.Stat(basePath); os.IsNotExist(pathErr) {
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return err
		}
	}

	if _, pathErr := os.Stat(i.Path()); !os.IsNotExist(pathErr) {
		return i.Update()
	}

	return i.Install()
}

// Update updates a repo.
func Update(i Installer) error {
	if _, pathErr := os.Stat(i.Path()); os.IsNotExist(pathErr) {
		return repo.ErrDoesNotExist
	}

	return i.Update()
}

// FindSource determines the correct Installer for the given source.
func FindSource(location string, home home.Home) (Installer, error) {
	installer, err := existingVCSRepo(location, home)
	if err != nil && err.Error() == "Cannot detect VCS" {
		return installer, repo.ErrMissingSource
	}
	return installer, err
}

// New determines and returns the correct Installer for the given source
func New(source, name, version string, home home.Home) (Installer, error) {
	if isLocalReference(source) {
		return NewLocalInstaller(source, name, home)
	}

	return NewVCSInstaller(source, name, version, home)
}

// isLocalReference checks if the source exists on the filesystem.
func isLocalReference(source string) bool {
	_, err := os.Stat(source)
	return err == nil
}

// isRepo checks if the directory contains a "bundles" directory.
func isRepo(dirname string) error {
	_, err := os.Stat(filepath.Join(dirname))
	if os.IsNotExist(err) {
		return repo.ErrDoesNotExist
	} else if err != nil {
		return err
	}
	_, err = os.Stat(filepath.Join(dirname, "bundles"))
	if os.IsNotExist(err) {
		return repo.ErrNotARepo
	}
	return err
}
