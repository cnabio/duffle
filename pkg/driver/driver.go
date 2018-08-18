package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

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
		// TODO: What would be great is if we could check for an executable
		// named `duffle-DRIVER` and wrap execution of that. I'm thinking we could
		// send it a JSON payload containing the Operation
		return &DebugDriver{}, errors.New("driver not found")
	}
}

// Operation describes the data passed into the driver to run an operation
type Operation struct {
	// Installation is the name of this installation
	Installation string
	// Action is the action to be performed
	Action string
	// Parameters are the paramaters to be injected into the container
	Parameters map[string]interface{}
	// Credentials are the credential sets to be loaded into the container
	Credentials []string
	// Image is the invocation image
	Image string
	// ImageType is the type of image.
	ImageType string
}

// Driver is capable of running a invocation image
type Driver interface {
	// Run executes the operation inside of the invocation image
	Run(*Operation) error
	// Handles receives an ImageType* and answers whether this driver supports that type
	Handles(string) bool
	// Config returns a map of configuration names and values that can be set via environment variable
	Config() map[string]string
	// SetConfig allows setting configuration, where name correspends to the key in Config, and value is
	// the value to be set.
	SetConfig(map[string]string)
}

// DebugDriver prints the information passed to a driver
//
// It does not ever run the image.
type DebugDriver struct {
	config map[string]string
}

func (d *DebugDriver) Run(op *Operation) error {
	data, err := json.MarshalIndent(op, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func (d *DebugDriver) Handles(dt string) bool {
	return true
}

func (d *DebugDriver) Config() map[string]string {
	return map[string]string{
		"VERBOSE": "Increase verbosity. true, false are supported values",
	}
}

func (d *DebugDriver) SetConfig(settings map[string]string) {
	d.config = settings
}

func isTrue(val string) bool {
	return strings.ToLower(val) == "true"
}
