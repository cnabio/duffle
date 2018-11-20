package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/image"
	"github.com/deis/duffle/pkg/repo"
)

func newBundlePushCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push BUNDLE",
		Short: "push a bundle to an image registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.Split(args[0], ":")
			if len(parts) != 2 {
				return fmt.Errorf("%s is of a wrong format, must be repo:tag", args[0])
			}
			name, version := parts[0], parts[1]
			h := home.Home(homePath())
			index, err := repo.LoadIndex(h.Repositories())
			if err != nil {
				return err
			}
			sha, err := index.Get(name, version)
			if err != nil {
				return err
			}
			fpath := filepath.Join(h.Bundles(), sha)
			data, err := ioutil.ReadFile(fpath)
			if err != nil {
				return err
			}
			cli := command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false)
			if err := cli.Initialize(flags.NewClientOptions()); err != nil {
				return err
			}
			digest, err := image.PushBundle(context.TODO(), cli, data, args[0])
			if err != nil {
				return err
			}
			fmt.Println("Digest of manifest is", digest)
			return nil
		},
	}
	return cmd
}
