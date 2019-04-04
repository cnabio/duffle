package main

import (
	"io"

	"github.com/spf13/cobra"
)

const bundleDesc = `
Manages bundles
`

func newBundleCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bundle",
		Aliases: []string{"bundles"},
		Short:   "manage bundles",
		Long:    bundleDesc,
	}
	cmd.AddCommand(
		newBundleListCmd(w),
		newInstallCmd(w),
		newBundleShowCmd(w),
		newBundleRemoveCmd(w),
	)
	return cmd
}
