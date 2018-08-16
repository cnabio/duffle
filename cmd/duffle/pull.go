package main

import (
	"io"

	"github.com/spf13/cobra"
)

// TODO
func newPullCmd(w io.Writer) *cobra.Command {
	const usage = `TODO`

	cmd := &cobra.Command{
		Use:   "pull",
		Short: usage,
		Long:  usage,
		Run: func(cmd *cobra.Command, args []string) {
			unimplemented("duffle pull")
		},
	}

	return cmd
}
