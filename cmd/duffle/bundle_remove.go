package main

import (
	"fmt"
	"io"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/repo"

	"github.com/spf13/cobra"
)

const bundleRemoveDesc = `Remove a bundle from the local storage.

This removes a bundle from the local storage so that it will no longer be locally
available. Remote bundles can be re-fetched with 'duffle pull'.
`

func newBundleRemoveCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove BUNDLE",
		Short: "remove a bundle from the local storage",
		Long:  bundleRemoveDesc,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			h := home.Home(homePath())
			index, err := repo.LoadIndex(h.Repositories())
			if err != nil {
				return err
			}

			bname := args[0]
			// TODO: Do we need to delete every record out of the local cache?
			if !index.Delete(bname) {
				fmt.Fprintf(w, "Bundle %q not found. Nothing deleted.", bname)
			}
			return index.WriteFile(h.Repositories(), 0644)
		},
	}
	return cmd
}
