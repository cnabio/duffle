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
		Use:     "key",
		Aliases: []string{"signature", "sig"},
		Short:   "manage keys",
		Long:    keyDesc,
	}
	cmd.AddCommand(
		newKeyAddCmd(w),
		newKeySignCmd(w),
		newKeyListCmd(w),
		newKeyVerifyCmd(w),
	)
	return cmd
}
