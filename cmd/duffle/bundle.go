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
		Use:   "bundle",
		Short: "manage bundles",
		Long:  bundleDesc,
	}
	cmd.AddCommand(
		newBundleListCmd(w),
		newBundleSignCmd(w),
		newBundleVerifyCmd(w),
	)
	return cmd
}
