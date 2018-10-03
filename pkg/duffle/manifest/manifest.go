package manifest

import (
	"os"
	"path/filepath"

	"github.com/deis/duffle/pkg/bundle"

	"github.com/technosophos/moniker"
)

// Manifest represents a duffle manifest.
type Manifest struct {
	Name        string                                `toml:"name,omitempty"`
	Components  map[string]*Component                 `toml:"components,omitempty"`
	Parameters  map[string]bundle.ParameterDefinition `toml:"parameters,omitempty"`
	Credentials map[string]bundle.CredentialLocation  `toml:"credentials,omitempty"`
}

// Component represents a component of a CNAB bundle
type Component struct {
	Name          string            `toml:"name,omitempty"`
	Builder       string            `toml:"builder,omitempty"`
	Configuration map[string]string `toml:"configuration,omitempty"`
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
