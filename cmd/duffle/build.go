package main

import (
	"io"

	"github.com/spf13/cobra"
)

// TODO
func newBuildCmd(w io.Writer) *cobra.Command {
	const usage = `TODO`

	cmd := &cobra.Command{
		Use:   "build",
		Short: usage,
		Long:  usage,
		Run: func(cmd *cobra.Command, args []string) {
			unimplemented("duffle build")
		},
	}

	return cmd
}
