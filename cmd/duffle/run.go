package main

import (
	"io"

	"github.com/spf13/cobra"
)

// TODO
func newRunCmd(w io.Writer) *cobra.Command {
	const usage = `TODO`

	cmd := &cobra.Command{
		Use:   "run",
		Short: usage,
		Long:  usage,
		Run: func(cmd *cobra.Command, args []string) {
			unimplemented("duffle run")
		},
	}

	return cmd
}
