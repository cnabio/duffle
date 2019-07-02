package manifest

import (
	"os"
	"path/filepath"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"

	"github.com/technosophos/moniker"
)

// Manifest represents a duffle manifest.
type Manifest struct {
	Name             string                       `json:"name" mapstructure:"name"`
	Version          string                       `json:"version" mapstructure:"version"`
	SchemaVersion    string                       `json:"schemaVersion" mapstructure:"schemaVersion"`
	Description      string                       `json:"description,omitempty" mapstructure:"description"`
	Keywords         []string                     `json:"keywords,omitempty" mapstructure:"keywords"`
	Maintainers      []bundle.Maintainer          `json:"maintainers,omitempty" mapstructure:"maintainers"`
	InvocationImages map[string]*InvocationImage  `json:"invocationImages,omitempty" mapstructure:"invocationImages"`
	Images           map[string]bundle.Image      `json:"images,omitempty" mapstructure:"images"`
	Actions          map[string]bundle.Action     `json:"actions,omitempty" mapstructure:"actions"`
	Parameters       *bundle.ParametersDefinition `json:"parameters,omitempty" mapstructure:"parameters"`
	Credentials      map[string]bundle.Credential `json:"credentials,omitempty" mapstructure:"credentials"`
	Definitions      definition.Definitions       `json:"definitions,omitempty" mapstructure:"definitions"`
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
