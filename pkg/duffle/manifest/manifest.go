package manifest

import (
	"os"
	"path/filepath"

	"github.com/deis/duffle/pkg/bundle"

	"github.com/technosophos/moniker"
)

// Manifest represents a duffle manifest.
type Manifest struct {
	Name        string                                `json:"name,omitempty"`
	Components  map[string]*Component                 `json:"components,omitempty"`
	Parameters  map[string]bundle.ParameterDefinition `json:"parameters,omitempty"`
	Credentials map[string]bundle.CredentialLocation  `json:"credentials,omitempty"`
}

// Component represents a component of a CNAB bundle
type Component struct {
	Name          string            `json:"name,omitempty"`
	Builder       string            `json:"builder,omitempty"`
	Configuration map[string]string `json:"configuration,omitempty"`
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
