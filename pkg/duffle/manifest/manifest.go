package manifest

import (
	"os"
	"path/filepath"

	"github.com/technosophos/moniker"
)

// Manifest represents a duffle.toml
type Manifest struct {
	Name       string   `toml:"name,omitempty"`
	Registry   string   `toml:"registry,omitempty"`
	Builder    string   `toml:"driver,omitempty"`
	Components []string `toml:"components,omitempty"`
}

// New creates a new manifest with the Environments intialized.
func New() *Manifest {
	return &Manifest{
		Name: generateName(),
	}
}

// generateName generates a name based on the current working directory or a random name.
func generateName() string {
	var name string
	cwd, err := os.Getwd()
	if err == nil {
		name = filepath.Base(cwd)
	} else {
		namer := moniker.New()
		name = namer.NameSep("-")
	}
	return name
}
