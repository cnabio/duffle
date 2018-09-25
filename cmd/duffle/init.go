package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/deis/duffle/pkg/duffle/home"

	"github.com/spf13/cobra"
)

const (
	initDesc = `
Initializes duffle with configuration required to start installing CNAB bundles.
`
)

type initCmd struct {
	dryRun bool
	w      io.Writer
}

func newInitCmd(w io.Writer) *cobra.Command {
	i := &initCmd{w: w}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "sets up local environment to work with duffle",
		Long:  initDesc,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.run()
		},
	}

	f := cmd.Flags()
	f.BoolVar(&i.dryRun, "dry-run", false, "go through all the steps without actually installing anything")

	return cmd
}

func (i *initCmd) run() error {
	home := home.Home(homePath())
	dirs := []string{
		home.String(),
		home.Logs(),
		home.Plugins(),
		home.Cache(),
		home.Claims(),
		home.Credentials(),
	}

	return i.ensureDirectories(dirs)
}

func (i *initCmd) ensureDirectories(dirs []string) error {
	fmt.Fprintln(i.w, "The following new directories will be created:")
	fmt.Fprintln(i.w, strings.Join(dirs, "\n"))
	for _, dir := range dirs {
		if fi, err := os.Stat(dir); err != nil {
			if !i.dryRun {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("Could not create %s: %s", dir, err)
				}
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", dir)
		}
	}
	return nil
}
