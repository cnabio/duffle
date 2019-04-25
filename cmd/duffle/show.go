package main

import (
	"io"

	"github.com/spf13/cobra"
)

func newShowCmd(w io.Writer) *cobra.Command {
	bsc := &bundleShowCmd{}
	bsc.w = w

	cmd := &cobra.Command{
		Use:   "show NAME",
		Short: bundleShowShortUsage,
		Long:  bsc.usage(false),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bsc.name = args[0]

			return bsc.run()
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&bsc.raw, "raw", "r", false, "Display the raw bundle manifest")

	return cmd
}
