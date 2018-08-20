package driver

import (
	"encoding/json"
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
		return &CommandDriver{Name: name}, nil
	}
}

// Operation describes the data passed into the driver to run an operation
type Operation struct {
	// Installation is the name of this installation
	Installation string `json:"installation_name"`
	// The revision ID for this installation
	Revision string `json:"revision"`
	// Action is the action to be performed
	Action string `json:"action"`
	// Parameters are the paramaters to be injected into the container
	Parameters map[string]interface{} `json:"parameters"`
	// Credentials are the credential sets to be loaded into the container
	Credentials []ResolvedCred `json:"credentials"`
	// Image is the invocation image
	Image string `json:"image"`
	// ImageType is the type of image.
	ImageType string `json:"image_type"`
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

type Configurable interface {
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
