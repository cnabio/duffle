package bundle

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

// ParseBuffer reads CNAB metadata out of a JSON byte stream
func ParseBuffer(data []byte) (Bundle, error) {
	b := Bundle{}
	err := json.Unmarshal(data, &b)
	return b, err
}

// Parse reads CNAB metadata from a JSON string
func Parse(text string) (Bundle, error) {
	return ParseBuffer([]byte(text))
}

// Parse reads CNAB metadata from a JSON string
func ParseReader(r io.Reader) (Bundle, error) {
	b := Bundle{}
	err := json.NewDecoder(r).Decode(&b)
	return b, err
}

func (b Bundle) WriteFile(dest string, mode os.FileMode) error {
	d, err := json.Marshal(b)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, d, mode)
}

// LocationRef specifies a location within the invocation package
type LocationRef struct {
	Path  string `json:"path" toml:"path"`
	Field string `json:"field" toml:"field"`
}

// Image describes a container image in the bundle
type Image struct {
	Name string        `json:"name" toml:"name"`
	URI  string        `json:"uri" toml:"uri"`
	Refs []LocationRef `json:"refs" toml:"refs"`
}

// InvocationImage contains the image type and location for the installation of a bundle
type InvocationImage struct {
	ImageType string `json:"imageType" toml:"imageType"`
	Image     string `json:"image" toml:"image"`
}

// CredentialLocation provides the location of a credential that the invocation
// image needs to use.
type CredentialLocation struct {
	Path                string `json:"path" toml:"path"`
	EnvironmentVariable string `json:"env" toml:"env"`
}

type Maintainer struct {
	// Name is a user name or organization name
	Name string `json:"name" toml:"name"`
	// Email is an optional email address to contact the named maintainer
	Email string `json:"email" toml:"email"`
	// Url is an optional URL to an address for the named maintainer
	URL string `json:"url" toml:"url"`
}

// Bundle is a CNAB metadata document
type Bundle struct {
	Name            string                         `json:"name" toml:"name"`
	Version         string                         `json:"version" toml:"version"`
	Description     string                         `json:"description" toml:"description"`
	Keywords        []string                       `json:"keywords" toml:"keywords"`
	Maintainers     []Maintainer                   `json:"maintainers" toml:"maintainers"`
	Deprecated      bool                           `json:"deprecated" toml:"deprecated"`
	InvocationImage InvocationImage                `json:"invocationImage" toml:"invocationImage"`
	Images          []Image                        `json:"images" toml:"images"`
	Parameters      map[string]ParameterDefinition `json:"parameters" toml:"parameters"`
	Credentials     map[string]CredentialLocation  `json:"credentials" toml:"credentials"`
}

// ValuesOrDefaults returns parameter values or the default parameter values
func ValuesOrDefaults(vals map[string]interface{}, b *Bundle) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	for name, def := range b.Parameters {
		if val, ok := vals[name]; ok {
			if err := def.ValidateParameterValue(val); err != nil {
				return res, err
			}
			res[name] = val
			continue
		}
		res[name] = def.DefaultValue
	}
	return res, nil
}
