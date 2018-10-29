package driver

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
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

func (d *DockerDriver) exec(op *Operation) error {

	// TODO - should ctx be passed here?
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		return fmt.Errorf("cannot create Docker client: %v", err)
	}

	if d.Simulate {
		return nil
	}

	// TODO - decide how to handle logs from Docker
	_, err = cli.ImagePull(ctx, op.Image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("cannot pull image %v: %v", op.Image, err)
	}

	var env []string
	for k, v := range op.Environment {
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}

	var mounts []mount.Mount

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
		// osPath is used to compute files to create
		// based on the operating system running duffle install
		osPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		base := filepath.Base(osPath)
		dir := filepath.Dir(osPath)

		// mount is the target mount path in the container
		//
		// TODO - make sure this is actually computed correctly
		// as the filepath.Dir(osPath) for any Unix path
		var m string
		x := strings.Split(path, `/`)
		if len(x) > 0 {
			m = strings.Join(x[:len(x)-1], "/")
		} else {
			m = strings.Join(x, "/")
		}

		// If it's another file in the same folder, add it to the existing tmp location
		if existingTmp, ok := tmpdirs[base]; ok {
			ioutil.WriteFile(filepath.Join(existingTmp, base), []byte(content), 0755)
			continue
		}

		tmp, err := ioutil.TempDir(os.TempDir(), "duffle-volume-")
		if err != nil {
			return err
		}
		tmpdirs[dir] = tmp
		localFile := filepath.Join(tmp, base)
		if err := ioutil.WriteFile(localFile, []byte(content), 0755); err != nil {
			fmt.Fprintln(op.Out, err)
		}
<<<<<<< HEAD

		mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: tmp, Target: m})
=======
		args = append(args, "--volume", fmt.Sprintf("%s:%s", localFile, fmt.Sprintf("%s/%s", dir, base)))
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
		fmt.Fprintln(op.Out, "--------> args")
		for _, arg := range args {
			fmt.Fprintln(op.Out, arg)
		}
		fmt.Fprintln(op.Out, "<-------- args")
>>>>>>> upstream/master
	}

	cfg := &container.Config{
		Image:      op.Image,
		Entrypoint: strslice.StrSlice{"/cnab/app/run"},
	}

	hostCfg := &container.HostConfig{Mounts: mounts}

	resp, err := cli.ContainerCreate(ctx, cfg, hostCfg, nil, "")
	if err != nil {
		return fmt.Errorf("cannot create container: %v", err)
	}
<<<<<<< HEAD
=======
	out, err := cmd.CombinedOutput()
	fmt.Fprintln(op.Out, "\n"+string(out)+"\n")
	return err
}
>>>>>>> upstream/master

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("cannot start container: %v", err)
	}

	statusc, errc := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errc:
		if err != nil {
			return fmt.Errorf("error in container: %v", err)
		}
	case <-statusc:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	fmt.Println(buf.String())

	return nil
}
