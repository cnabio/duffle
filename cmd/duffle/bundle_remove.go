package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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
		Use:     "remove BUNDLE",
		Aliases: []string{"rm"},
		Short:   "remove a bundle from the local storage",
		Long:    bundleRemoveDesc,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			h := home.Home(homePath())
			index, err := repo.LoadIndex(h.Repositories())
			if err != nil {
				return err
			}

			bname := args[0]

			vers, ok := index.GetVersions(bname)
			if !ok {
				fmt.Fprintf(w, "Bundle %q not found. Nothing deleted.", bname)
				return nil
			}

			if !index.Delete(bname) {
				fmt.Fprintf(w, "Bundle %q not found. Nothing deleted.", bname)
				return nil
			}
			if err := index.WriteFile(h.Repositories(), 0644); err != nil {
				return err
			}

			// Now that the index record is gone, we can safely delete records.
			// It is odd that there is no library func to do this.
			for _, sha := range vers {
				fpath := filepath.Join(h.Bundles(), sha)
				if err := os.Remove(fpath); err != nil {
					fmt.Fprintf(w, "WARNING: could not delete stake record %q", fpath)
				}
			}
			return nil
		},
	}
	return cmd
}
