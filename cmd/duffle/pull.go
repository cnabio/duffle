package main

import (
	"errors"
	"fmt"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("This command requires at least one argument: BUNDLE (CNAB bundle name)\nValid inputs:\n\t$ duffle pull BUNDLE")
			}
			bundleFile, err := findBundleJSON(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(w, bundleFile)
			return nil
		},
	}

	return cmd
}
