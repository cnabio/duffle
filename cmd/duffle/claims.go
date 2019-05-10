package main

import (
	"io"

	"github.com/spf13/cobra"
)

const claimsDesc = `
Work with claims and existing releases.

A claim is a record of a release. When a bundle is installed, Duffle retains a
claim that tracks that release. Subsequent operations (like upgrades) will
modify the claim record.

The claim tools provide features for working directly with claims.

The default claim storage system stores claims in $HOME/.duffle/claims. Other drivers
can be configured by setting the DUFFLE_CLAIM_STORAGE environment variable. Known
values are:

  - fs, filesystem: Use the filesystem (default)
  - mongodb: Use MongoDB-compatible storage. You may want to set DUFFLE_MONGODB_URL.

  Configuring MongoDB
  ===================

  To use MongoDB you may need to provide additional information via environment
  variables.

  DUFFLE_MONGODB_URL: The URL to the mongo database. Default is "mongodb://localhost:27017/duffle"
`

func newClaimsCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "claims",
		Short:   "manage claims",
		Long:    claimsDesc,
		Aliases: []string{"claim"},
	}

	cmd.AddCommand(newClaimsShowCmd(w))
	cmd.AddCommand(newClaimListCmd(w))

	return cmd
}
