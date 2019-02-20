package driver

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/credentials"
)

// Input describes the input configuration available on stdin from within the invocation image
type Input struct {
	// Installation is the details of the installation
	Installation InputInstallation `json:"installation"`
	// InvocationImage is the details of the image that is being invoked
	InvocationImage bundle.InvocationImage `json:"invocation_image"`
	// Bundle is the entire contents of the bundle.json
	Bundle bundle.Bundle `json:"bundle"`
	// Parameters is the set of resolved parameters
	Parameters map[string]interface{} `json:"parameters"`
	// Credentials is the set of resolved credentials
	Credentials credentials.Set `json:"credentials"`
}

// InputInstallation describes the name, action, revision for the installation
type InputInstallation struct {
	// Name is the name of the installation
	Name string `json:"name"`
	// Action is the action to be performed
	Action string `json:"action"`
	// Revision is the revision ID for this installation
	Revision string `json:"revision"`
}

// ImageType constants provide some of the image types supported
// TODO: I think we can remove all but Docker, since the rest are supported externally
const (
	ImageTypeDocker = "docker"
	ImageTypeOCI    = "oci"
	ImageTypeQCOW   = "qcow"
)

// Lookup takes a driver name and tries to resolve the most pertinent driver.
func Lookup(name string) (Driver, error) {
	switch name {
	case "docker":
		return &DockerDriver{}, nil
	case "debug":
		return &DebugDriver{}, nil
	default:
		return &CommandDriver{Name: name}, nil
	}
}

// Operation describes the data passed into the driver to run an operation
type Operation struct {
	// Image is the invocation image
	Image string `json:"image"`
	// ImageType is the type of image.
	ImageType string `json:"image_type"`
	// Output stream for log messages from the driver
	Out io.Writer

	Input Input `json:"input"`
}

// ResolvedCred is a credential that has been resolved and is ready for injection into the runtime.
type ResolvedCred struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Driver is capable of running a invocation image
type Driver interface {
	// Run executes the operation inside of the invocation image
	Run(*Operation) error
	// Handles receives an ImageType* and answers whether this driver supports that type
	Handles(string) bool
}

// Configurable drivers can explain their configuration, and have it explicitly set
type Configurable interface {
	// Config returns a map of configuration names and values that can be set via environment variable
	Config() map[string]string
	// SetConfig allows setting configuration, where name corresponds to the key in Config, and value is
	// the value to be set.
	SetConfig(map[string]string)
}

// DebugDriver prints the information passed to a driver
//
// It does not ever run the image.
type DebugDriver struct {
	config map[string]string
}

// Run executes the operation on the Debug  driver
func (d *DebugDriver) Run(op *Operation) error {
	data, err := json.MarshalIndent(op, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(op.Out, string(data))
	return nil
}

// Handles always returns true, effectively claiming to work for any image type
func (d *DebugDriver) Handles(dt string) bool {
	return true
}

// Config returns the configuration help text
func (d *DebugDriver) Config() map[string]string {
	return map[string]string{
		"VERBOSE": "Increase verbosity. true, false are supported values",
	}
}

// SetConfig sets configuration for this driver
func (d *DebugDriver) SetConfig(settings map[string]string) {
	d.config = settings
}
