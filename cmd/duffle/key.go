package main

import (
	"io"

	"github.com/spf13/cobra"
)

const keyDesc = `
Manages OpenPGP keys, signatures, and attestations.
`

// TODO
func newKeyCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "key",
		Aliases: []string{"keys"},
		Short:   "manage keys",
		Long:    keyDesc,
	}
	cmd.AddCommand(
		newKeyAddCmd(w),
		newKeyListCmd(w),
	)
	return cmd
}
