package manifest

import (
	"os"
	"path/filepath"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"

	"github.com/technosophos/moniker"
)

// Manifest represents a duffle manifest.
type Manifest struct {
	Name               string                       `json:"name"`
	Version            string                       `json:"version"`
	SchemaVersion      string                       `json:"schemaVersion"`
	Description        string                       `json:"description,omitempty"`
	Keywords           []string                     `json:"keywords,omitempty"`
	Maintainers        []bundle.Maintainer          `json:"maintainers,omitempty"`
	InvocationImages   map[string]*InvocationImage  `json:"invocationImages,omitempty"`
	Images             map[string]bundle.Image      `json:"images,omitempty"`
	Actions            map[string]bundle.Action     `json:"actions,omitempty"`
	Parameters         map[string]bundle.Parameter  `json:"parameters,omitempty"`
	Credentials        map[string]bundle.Credential `json:"credentials,omitempty"`
	Definitions        definition.Definitions       `json:"definitions,omitempty"`
	Outputs            map[string]bundle.Output     `json:"outputs,omitempty"`
	Custom             map[string]interface{}       `json:"custom,omitempty"`
	License            string                       `json:"license,omitempty"`
	RequiredExtensions []string                     `json:"requiredExtensions,omitempty"`
}

// InvocationImage represents an invocation image component of a CNAB bundle
type InvocationImage struct {
	Name          string            `json:"name"`
	Builder       string            `json:"builder"`
	Configuration map[string]string `json:"configuration"`
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
