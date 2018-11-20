package config

import (
	"fmt"
	"os"

	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"

	"github.com/spf13/afero"
)

type Context struct {
	DockerClient *command.DockerCli //todo: try and narrow this down so it's mockable
	FileSystem   *afero.Afero
}

func New() (*Context, error) {
	cli := command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false)
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return nil, err
	}
	if cli == nil {
		fmt.Printf("well this is unfortunate")
	}
	return &Context{
		DockerClient: cli,
		FileSystem:   &afero.Afero{Fs: afero.NewOsFs()},
	}, nil
}
