package driver

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// DockerDriver is capable of running Docker invocation images using Docker itself.
type DockerDriver struct {
	config map[string]string
	// If true, this will not actually run Docker
	Simulate bool
}

// Run executes the Docker driver
func (d *DockerDriver) Run(op *Operation) error {
	return d.exec(op)
}

// Handles indicates that the Docker driver supports "docker" and "oci"
func (d *DockerDriver) Handles(dt string) bool {
	return dt == ImageTypeDocker || dt == ImageTypeOCI
}

func (d *DockerDriver) exec(op *Operation) error {
	img := op.Image
	env := op.Environment
	// FIXME: This is all temporary code. We should really just link the Docker library and
	// directly send this.
	args := []string{"run"}
	for k, v := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%v", k, v))
	}

	// To pass secrets into the running container, we loop through all of the files
	// and store them in a local temp directory (one file per directory). Then we
	// mount all of the directories to Docker.
	tmpdirs := map[string]string{}
	defer func() {
		for _, tmp := range tmpdirs {
			os.RemoveAll(tmp)
		}
	}()
	for path, content := range op.Files {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		dir := filepath.Dir(path)

		// If it's another file in the same folder, add it to the existing tmp location
		if existingTmp, ok := tmpdirs[base]; ok {
			ioutil.WriteFile(filepath.Join(existingTmp, base), []byte(content), 0755)
			continue
		}

		// FIXME: On Docker-for-Mac, you must hard code the temp dir location. But
		// does this work on Windows?
		tmp, err := ioutil.TempDir("/tmp", "duffle-volume-")
		if err != nil {
			return err
		}
		tmpdirs[dir] = tmp
		localFile := filepath.Join(tmp, base)
		if err := ioutil.WriteFile(localFile, []byte(content), 0755); err != nil {
			fmt.Println(err)
		}
		args = append(args, "--volume", fmt.Sprintf("%s:%s", tmp, dir))
	}

	// TODO: For now, we mount the docker socket to alllow things like Compose
	// to run inside of a CNAB bundle. This should be configurable.
	// See https://github.com/docker/compose/blob/master/script/run/run.sh
	// Also https://media.giphy.com/media/RIECDaCdxqKha/giphy.gif
	args = append(args, "--volume", "/var/run/docker.sock:/var/run/docker.sock")

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
	fmt.Println("\n" + string(out) + "\n")
	return err
}

// Config returns the Docker driver configuration options
func (d *DockerDriver) Config() map[string]string {
	return map[string]string{
		"VERBOSE": "Increase verbosity. true, false are supported values",
	}
}

// SetConfig sets Docker driver configuration
func (d *DockerDriver) SetConfig(settings map[string]string) {
	d.config = settings
}
