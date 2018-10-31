package driver

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	unix_path "path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
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
	err = client.FromEnv(cli)
	if err != nil {
		return fmt.Errorf("cannot update Docker client: %v", err)
	}
	if d.Simulate {
		return nil
	}
	fmt.Println("Pulling Invocation Image...")
	pullReader, err := cli.ImagePull(ctx, op.Image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("cannot pull image %v: %v", op.Image, err)
	}
	termFd, isTerm := term.GetFdInfo(os.Stdout)
	err = jsonmessage.DisplayJSONMessagesStream(pullReader, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		return err
	}
	var env []string
	for k, v := range op.Environment {
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: "/var/run/docker.sock",
			Target: "/var/run/docker.sock"},
	}
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

	for path, content := range op.Files {

		if !unix_path.IsAbs(path) {
			return errors.New("destination path should be an absolute unix path")
		}
		tarContent, err := generateTar(path, content)
		if err != nil {
			return fmt.Errorf("error staging files for %s", path)
		}
		options := types.CopyToContainerOptions{
			AllowOverwriteDirWithFile: false,
		}
		// This copies the tar to the root of the container. The tar has been assembled using the
		// path from the given file, starting at the /.
		err = cli.CopyToContainer(ctx, resp.ID, "/", tarContent, options)
		if err != nil {
			return fmt.Errorf("error copying %s to / in container: %s", path, err)
		}
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
			if err != nil {
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

func generateTar(dst string, content string) (io.Reader, error) {
	r, w := io.Pipe()
	tw := tar.NewWriter(w)
	go func() {
		hdr := &tar.Header{
			Name: dst,
			Mode: 0644,
			Size: int64(len(content)),
		}
		tw.WriteHeader(hdr)
		tw.Write([]byte(content))
		w.Close()
	}()
	return r, nil
}
