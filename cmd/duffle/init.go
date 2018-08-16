package main

import (
	"io"

	"github.com/spf13/cobra"
)

// TODO
func newInitCmd(w io.Writer) *cobra.Command {
	const usage = `TODO`

	cmd := &cobra.Command{
		Use:   "init",
		Short: usage,
		Long:  usage,
		Run: func(cmd *cobra.Command, args []string) {
			unimplemented("duffle init")
		},
	}

	return cmd
}
