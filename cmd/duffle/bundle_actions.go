package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

const bundleActionsDesc = ` Lists all actions available in a bundle`

type bundleActionsCmd struct {
	out        io.Writer
	bundleFile string
	insecure   bool
}

func newBundleActionsCmd(w io.Writer) *cobra.Command {
	a := &bundleActionsCmd{out: w}

	cmd := &cobra.Command{
		Use:   "actions BUNDLE",
		Short: bundleActionsDesc,
		Long:  bundleActionsDesc,
		RunE: func(cmd *cobra.Command, args []string) error {

			bundle, err := bundleFileOrArg1(args, a.bundleFile)
			if err != nil {
				return err
			}
			return a.run(bundle)
		},
	}

	cmd.Flags().StringVarP(&a.bundleFile, "file", "f", "", "path to the bundle file to show actions for")
	cmd.Flags().BoolVarP(&a.insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")

	return cmd
}

func (a *bundleActionsCmd) run(bundleFile string) error {

	// Verify that file exists
	if fi, err := os.Stat(bundleFile); err != nil {
		return fmt.Errorf("cannot find bundle file to sign: %v", err)
	} else if fi.IsDir() {
		return errors.New("cannot sign a directory")
	}

	bun, err := loadBundle(bundleFile, a.insecure)
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
