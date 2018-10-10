package main

import (
	"io"

	"github.com/spf13/cobra"
)

const keyDesc = `
Manage OpenPGP keys, signatures, and attestations.
`

// TODO
func newKeyCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "manage keys",
		Long:  keyDesc,
	}
	cmd.AddCommand(
		newKeySignCmd(w),
		newKeyListCmd(w),
	)
	return cmd
}
