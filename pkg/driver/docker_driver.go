package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	unix_path "path"
	"path/filepath"
	"strings"

	"github.com/deis/duffle/pkg/duffle/home"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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
	//err = client.FromEnv(cli)
	//if err != nil {
	//		return fmt.Errorf("cannot update Docker client: %v, err")
	//	}
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
		if !unix_path.IsAbs(path) {
			return errors.New("destination path should be an absolute unix path")
		}
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

		tmpDirRoot := home.DefaultHome()
		tmp, err := ioutil.TempDir(tmpDirRoot, "duffle-volume-")
		if err != nil {
			return err
		}
		tmpdirs[dir] = tmp
		localFile := filepath.Join(tmp, base)
		if err := ioutil.WriteFile(localFile, []byte(content), 0777); err != nil {
			fmt.Fprintln(op.Out, err)
			return err
		}

		mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: tmp, Target: m})
	}

	// mounts = append(mounts, mount.Mount{
	// 	Type:   mount.TypeBind,
	// 	Source: "/var/run/docker.sock",
	// 	Target: "/var/run/docker.sock"},
	// )

	cfg := &container.Config{
		Image:      op.Image,
		Env:        env,
		Entrypoint: strslice.StrSlice{"/cnab/app/run"},
	}

	hostCfg := &container.HostConfig{Mounts: mounts, AutoRemove: true}

	resp, err := cli.ContainerCreate(ctx, cfg, hostCfg, nil, "")
	if err != nil {
		return fmt.Errorf("cannot create container: %v", err)
	}

	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("cannot start container: %v", err)
	}

	attach, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		return fmt.Errorf("unable to retrieve logs: %v", err)
	}
	go func() {
		defer attach.Close()
		for {
			_, err := stdcopy.StdCopy(os.Stdout, os.Stderr, attach.Reader)
			if err == io.EOF {
				break
			}
		}
	}()

	statusc, errc := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errc:
		if err != nil {
			return fmt.Errorf("error in container: %v", err)
		}
	case <-statusc:
	}
	return err
}
