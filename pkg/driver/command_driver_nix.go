// +build !windows

package driver

import (
	"fmt"
	"os"
	"os/exec"
)

// CheckDriverExists checks to see if the named driver exists
func (d *CommandDriver) CheckDriverExists() bool {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("command -v %s", d.cliName()))
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}
