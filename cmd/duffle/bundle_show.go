package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

const bundleShowShortUsage = `return low-level information on application bundles`

type bundleShowCmd struct {
	name string
	raw  bool
	w    io.Writer
}

func newBundleShowCmd(w io.Writer) *cobra.Command {
	bsc := &bundleShowCmd{}
	bsc.w = w

	cmd := &cobra.Command{
		Use:   "show NAME",
		Short: bundleShowShortUsage,
		Long:  bsc.usage(true),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bsc.name = args[0]

			return bsc.run()
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&bsc.raw, "raw", "r", false, "Display the raw bundle manifest")

	return cmd
}

func (bsc *bundleShowCmd) usage(bundleSubCommand bool) string {
	commandName := "show"
	if bundleSubCommand {
		commandName = "bundle show"
	}

	return fmt.Sprintf(` Returns information about an application bundle.

	Example:
		$ duffle %s duffle/example:0.1.0

`, commandName)
}

func (bsc *bundleShowCmd) run() error {
	bundleFile, err := getBundleFilepath(bsc.name, homePath())
	if err != nil {
		return err
	}

	if bsc.raw {
		f, err := os.Open(bundleFile)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(bsc.w, f)
		return err
	}

	bun, err := loadBundle(bundleFile)
	if err != nil {
		return err
	}

	_, err = bun.WriteTo(bsc.w)

	return err
}
