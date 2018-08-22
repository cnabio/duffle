package main

import (
	"fmt"
	"io"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/utils/crud"
	"github.com/spf13/cobra"
)

func newListCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list installed apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: unify with similar code in duffle install currently awaiting PR
			claimsPath := home.Home(homePath()).Claims()
			claimsStore := claim.NewClaimStore(crud.NewFileSystemStore(claimsPath, "json"))
			claims, err := claimsStore.List()
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
