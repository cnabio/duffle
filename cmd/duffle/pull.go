package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newPullCmd(w io.Writer) *cobra.Command {
	const usage = `pulls a CNAB bundle from a repository

This command pulls a CNAB bundle into the cache without installing it.

Example:
	$ duffle pull duffle/example:0.1.0
`

	cmd := &cobra.Command{
		Use:   "pull",
		Short: usage,
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := getBundleFile(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(w, path)
			return nil
		},
	}

	return cmd
}
