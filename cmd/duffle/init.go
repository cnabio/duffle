package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cnabio/duffle/pkg/duffle/home"
	"github.com/cnabio/duffle/pkg/ohai"
)

const (
	initDesc = `
Explicitly control the creation of the Duffle environment.

Normally, Duffle initializes itself. But on occasion, you may wish to customize Duffle's initialization,
 or testing to see what directories will be created. This command is provided for such a reason.

This command will create a subdirectory in your home directory, and use that directory for storing
configuration, preferences, and persistent data.
`
)

type initCmd struct {
	dryRun  bool
	w       io.Writer
	verbose bool
}

func newInitCmd(w io.Writer) *cobra.Command {
	i := &initCmd{w: w, verbose: true}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "set up local environment to work with duffle",
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

// autoInit is called by the root command for all calls except init and version.
func autoInit(w io.Writer, verbose bool) error {
	i := initCmd{
		w:       w,
		verbose: verbose,
	}
	return i.run()
}

func (i *initCmd) run() error {
	home := home.Home(homePath())
	dirs := []string{
		home.String(),
		home.Bundles(),
		home.Logs(),
		home.Plugins(),
		home.Claims(),
		home.Credentials(),
	}

	files := []string{
		home.Repositories(),
	}

	if i.verbose {
		ohai.Fohailn(i.w, "The following new directories will be created:")
		fmt.Fprintln(i.w, strings.Join(dirs, "\n"))
	}

	if !i.dryRun {
		if err := ensureDirectories(dirs); err != nil {
			return err
		}
	}

	if i.verbose {
		ohai.Fohailn(i.w, "The following new files will be created:")
		fmt.Fprintln(i.w, strings.Join(files, "\n"))
	}

	if !i.dryRun {
		if err := ensureFiles(files); err != nil {
			return err
		}
	}

	return nil
}

func ensureDirectories(dirs []string) error {
	for _, dir := range dirs {
		if fi, err := os.Stat(dir); err != nil {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("Could not create %s: %s", dir, err)
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", dir)
		}
	}
	return nil
}

func ensureFiles(files []string) error {
	for _, name := range files {
		f, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}
