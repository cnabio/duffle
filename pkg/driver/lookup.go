package driver

import (
	"fmt"

	"github.com/deislabs/cnab-go/driver"
)

// Lookup takes a driver name and tries to resolve the most pertinent driver.
func Lookup(name string) (driver.Driver, error) {
	switch name {
	case "docker":
		return &DockerDriver{}, nil
	case "kubernetes", "k8s":
		return &KubernetesDriver{}, nil
	case "debug":
		return &driver.DebugDriver{}, nil
	default:
		cmddriver := &CommandDriver{Name: name}
		if cmddriver.CheckDriverExists() {
			return cmddriver, nil
		}

		return nil, fmt.Errorf("unsupported driver or driver not found in PATH: %s", name)
	}
}
