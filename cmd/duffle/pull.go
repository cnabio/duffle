package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newPullCmd(w io.Writer) *cobra.Command {
	const usage = `Pulls a CNAB bundle into the cache without installing it.

Example:
	$ duffle pull duffle/example:0.1.0
`

	var insecure bool
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "pull a CNAB bundle from a repository",
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := getBundleFile(args[0], insecure)
			if err != nil {
				return err
			}
			fmt.Fprintln(w, path)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")

	return cmd
}
