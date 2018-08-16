package main

import (
	"io"

	"github.com/spf13/cobra"
)

// TODO
func newPushCmd(w io.Writer) *cobra.Command {
	const usage = `TODO`

	cmd := &cobra.Command{
		Use:   "push",
		Short: usage,
		Long:  usage,
		Run: func(cmd *cobra.Command, args []string) {
			unimplemented("duffle push")
		},
	}

	return cmd
}
