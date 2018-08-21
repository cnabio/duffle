package driver

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DockerDriver is capable of running Docker invocation images using Docker itself.
type DockerDriver struct {
	config map[string]string
	// If true, this will not actually run Docker
	Simulate bool
}

func (d *DockerDriver) Run(op *Operation) error {
	env := map[string]interface{}{
		"CNAB_INSTALLATION_NAME": op.Installation,
		"CNAB_ACTION":            op.Action,
		"CNAB_BUNDLE_NAME":       op.Image,
	}
	for k, v := range op.Parameters {
		// TODO: Vet against bundle's parameters.json
		env[fmt.Sprintf("CNAB_P_%s", strings.ToUpper(k))] = v
	}

	return d.exec(op.Image, env)
}

func (d *DockerDriver) Handles(dt string) bool {
	return dt == ImageTypeDocker || dt == ImageTypeOCI
}

func (d *DockerDriver) exec(img string, env map[string]interface{}) error {
	// FIXME: This is all temporary code. We should really just link the Docker library and
	// directly send this.
	args := []string{"run"}
	for k, v := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%v", k, v))
	}
	// TODO: Should we hard code in the call to run? This might actually make it possible
	// for CNAB devs to create a default command that is perhaps user-oriented (like setting
	// the default command to help text).
	args = append(args, img, "/cnab/app/run")

	if isTrue(d.config["VERBOSE"]) {
		fmt.Println("--------> args")
		for _, arg := range args {
			fmt.Println(arg)
		}
		fmt.Println("<-------- args")
	}

	if d.Simulate {
		return nil
	}
	var err error
	cmd := exec.Command("docker", args...)
	cmd.Dir, err = os.Getwd()
	if err != nil {
		return err
	}
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	return err
}

func (d *DockerDriver) Config() map[string]string {
	return map[string]string{
		"VERBOSE": "Increase verbosity. true, false are supported values",
	}
}

func (d *DockerDriver) SetConfig(settings map[string]string) {
	d.config = settings
}
