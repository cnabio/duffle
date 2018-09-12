package main

import (
	"io"

	"github.com/spf13/cobra"
)

const credentialDesc = `
Manage credential sets
`

func newCredentialsCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "credentials",
		Short:   "manage credential sets",
		Long:    credentialDesc,
		Aliases: []string{"creds", "credential"},
	}

	cmd.AddCommand(
		newCredentialListCmd(w),
		newCredentialRemoveCmd(w),
		newCredentialAddCmd(w),
		newCredentialShowCmd(w),
	)

	return cmd
}
