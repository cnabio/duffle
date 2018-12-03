package manifest

import (
	"os"
	"path/filepath"

	"github.com/deislabs/duffle/pkg/bundle"

	"github.com/technosophos/moniker"
)

// Manifest represents a duffle manifest.
type Manifest struct {
	Name             string                                `json:"name" mapstructure:"name"`
	Version          string                                `json:"version" mapstructure:"version"`
	Description      string                                `json:"description" mapstructure:"description"`
	Keywords         []string                              `json:"keywords" mapstructure:"keywords"`
	Maintainers      []bundle.Maintainer                   `json:"maintainers" mapstructure:"maintainers"`
	InvocationImages map[string]*InvocationImage           `json:"invocationImages" mapstructure:"invocationImages"`
	Parameters       map[string]bundle.ParameterDefinition `json:"parameters" mapstructure:"parameters"`
	Credentials      map[string]bundle.Location            `json:"credentials" mapstructure:"credentials"`
}

// InvocationImage represents an invocation image component of a CNAB bundle
type InvocationImage struct {
	Name          string            `json:"name" mapstructure:"name"`
	Builder       string            `json:"builder" mapstructure:"builder"`
	Configuration map[string]string `json:"configuration" mapstructure:"configuration"`
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
