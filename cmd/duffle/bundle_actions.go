package main

import (
	"fmt"
	"io"

	"github.com/deislabs/duffle/pkg/duffle/home"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

const bundleActionsDesc = `list all actions available in a bundle`

type bundleActionsCmd struct {
	out          io.Writer
	home         home.Home
	bundle       string
	bundleIsFile bool
}

func newBundleActionsCmd(w io.Writer) *cobra.Command {
	a := &bundleActionsCmd{out: w}

	cmd := &cobra.Command{
		Use:   "actions BUNDLE",
		Short: bundleActionsDesc,
		Long:  bundleActionsDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			a.bundle = args[0]
			a.home = home.Home(homePath())

			return a.run()
		},
	}

	cmd.Flags().BoolVarP(&a.bundleIsFile, "bundle-is-file", "f", false, "Indicates that the bundle source is a file path")

	return cmd
}

func (a *bundleActionsCmd) run() error {
	bundleFile, err := resolveBundleFilePath(a.bundle, a.home.String(), a.bundleIsFile)
	if err != nil {
		return err
	}

	bun, err := loadBundle(bundleFile)
	if err != nil {
		return err
	}

	table := uitable.New()
	table.MaxColWidth = 100
	table.Wrap = true

	table.AddRow("ACTION", "STATELESS", "MODIFIES", "DESCRIPTION")
	for name, act := range bun.Actions {
		table.AddRow(name, act.Stateless, act.Modifies, act.Description)
	}

	fmt.Fprintln(a.out, table)

	return nil
}
