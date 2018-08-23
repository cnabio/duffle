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
		home.Repositories(),
		home.Claims(),
		home.Credentials(),
	}

	// The only way I can get linter to pass is by commenting this out and simplifying
	// See issue #113
	return i.ensureDirectories(dirs)
	/*
		if err := i.ensureDirectories(dirs); err != nil {
			return err
		}
		// TODO: add repos here
		// return i.ensureRepositories()
		return nil
	*/
}

// ensureRepositories checks to see if the default repositories exists.
//
// If the repo does not exist, this function will create it.
/* See #115 - commented out to appease angry linter
func (i *initCmd) ensureRepositories() error {
	ohai.Fohailn(i.w, "Installing default repositories...")

	// TODO: add repos here
	addArgs := []string{}

	repoCmd, _, err := rootCmd.Find([]string{"repo", "add"})
	if err != nil {
		return err
	}
	if i.dryRun {
		return nil
	}
	return repoCmd.RunE(repoCmd, addArgs)
}
*/

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
