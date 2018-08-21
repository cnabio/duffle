package main

import (
	"io"

	"github.com/spf13/cobra"
)

const repoDesc = `
Manage repositories.
`

// TODO
func newRepoCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "manage repositories",
		Long:  repoDesc,
	}
	cmd.AddCommand(
		newRepoAddCmd(w),
		newRepoListCmd(w),
		newRepoRemoveCmd(w),
		newRepoUpdateCmd(w),
	)
	return cmd
}
