package credentials

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/deis/duffle/pkg/bundle"

	yaml "gopkg.in/yaml.v2"
)

// Set is an actual set of resolved credentials.
// This is the output of resolving a credentialset file.
type Set map[string]Destination

// Flatten returns a map of env vars and a map of files.
func (s Set) Flatten() (env, files map[string]string) {
	// We want empty maps, not nil maps
	env, files = map[string]string{}, map[string]string{}
	for _, dest := range s {
		if dest.EnvVar != "" {
			env[dest.EnvVar] = dest.Value
		}
		if dest.Path != "" {
			files[dest.Path] = dest.Value
		}
	}
	return
}

// CredentialSet represents a collection of credentials
type CredentialSet struct {
	// Name is the name of the credentialset.
	Name string
	// Creadentials is a list of credential specs.
	Credentials []CredentialStrategy
}

// Load a CredentialSet from a file at a given path.
func Load(path string) (*CredentialSet, error) {
	cset := &CredentialSet{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return cset, err
	}
	return cset, yaml.Unmarshal(data, cset)
}

// Validate compares the given credentials with the spec.
//
// This will result in an error only if:
// - a parameter in the spec is not present in the given set
// - a parameter in the given set does not match the format required by the spec
//
// It is allowed for spec to specify both an env var and a file. In such case, if
// the givn set provides either, it will be considered valid.
func Validate(given Set, spec map[string]bundle.CredentialLocation) error {
	for name, loc := range spec {
		if !isValidCred(given, loc) {
			return fmt.Errorf("bundle requires credential for %s", name)
		}
	}
	return nil
}

func isValidCred(given Set, loc bundle.CredentialLocation) bool {
	for _, v := range given {
		if loc.EnvironmentVariable == v.EnvVar {
			return true
		}
		if loc.Path == v.Path {
			return true
		}
	}
	return false
}

// Resolve looks up the credentials as described in Source, then copies the resulting value into Destination.
func (c *CredentialSet) Resolve() (Set, error) {
	l := len(c.Credentials)
	res := make(map[string]Destination, l)
	for i := 0; i < l; i++ {
		src := c.Credentials[i].Source
		dest := c.Credentials[i].Destination
		// Precedence is Command, Path, EnvVar, Value
		switch {
		case src.Command != "":
			data, err := execCmd(src.Command)
			if err != nil {
				return res, err
			}
			dest.Value = string(data)
		case src.Path != "":
			data, err := ioutil.ReadFile(os.ExpandEnv(src.Path))
			if err != nil {
				return res, fmt.Errorf("credential %q: %s", c.Credentials[i].Name, err)
			}
			dest.Value = string(data)
		case src.EnvVar != "":
			var ok bool
			dest.Value, ok = os.LookupEnv(src.EnvVar)
			if ok {
				break
			}
			fallthrough
		default:
			dest.Value = src.Value
		}
		res[c.Credentials[i].Name] = dest
	}
	return res, nil
}

func execCmd(cmd string) ([]byte, error) {
	parts := strings.Split(cmd, " ")
	c := parts[0]
	args := parts[1:]
	run := exec.Command(c, args...)

	return run.CombinedOutput()
}

// CredentialStrategy represents a source credential and the destination to which it should be sent.
type CredentialStrategy struct {
	Name        string      `json:"name" yaml:"name"`
	Source      Source      `json:"source,omitempty" yaml:"source,omitempty"`
	Destination Destination `json:"destination" yaml:"destination"`
}

// Source represents a strategy for loading a credential from local host.
type Source struct {
	//Type    string `json:"type"`
	Path    string `json:"path,omitempty" yaml:"path,omitempty"`
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
	Value   string `json:"value,omitempty" yaml:"value,omitempty"`
	EnvVar  string `json:"env,omitempty" yaml:"env,omitempty"`
}

// Destination reprents a strategy for injecting a credential into an image.
type Destination struct {
	Path   string `json:"path,omitempty" yaml:"path,omitempty"`
	EnvVar string `json:"env,omitempty" yaml:"env,omitempty"`
	Value  string `json:"value,omitempty" yaml:"value,omitempty"`
}
