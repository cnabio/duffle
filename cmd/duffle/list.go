package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newListCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list installed apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			claims, err := claimStorage().List()
			if err != nil {
				return err
			}

			for _, claim := range claims {
				fmt.Fprintln(w, claim)
			}
			return nil
		},
	}
	return cmd
}
