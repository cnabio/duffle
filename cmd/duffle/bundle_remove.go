package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/repo"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
)

const bundleRemoveDesc = `Remove a bundle from the local storage.

This removes a bundle from the local storage so that it will no longer be locally
available. Bundles can be rebuilt with 'duffle build'.

If a SemVer range is provided with '--version'/'-r' then only releases that match
that range will be removed.
`

func newBundleRemoveCmd(w io.Writer) *cobra.Command {
	var versions string
	cmd := &cobra.Command{
		Use:     "remove BUNDLE",
		Aliases: []string{"rm"},
		Short:   "remove a bundle from the local storage",
		Long:    bundleRemoveDesc,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bname := args[0]

			h := home.Home(homePath())
			index, err := repo.LoadIndex(h.Repositories())
			if err != nil {
				return err
			}

			vers, ok := index.GetVersions(bname)
			if !ok {
				fmt.Fprintf(w, "Bundle %q not found. Nothing deleted.", bname)
				return nil
			}

			// If versions is set, we short circuit and only delete specific versions.
			if versions != "" {
				fmt.Fprintln(w, "Only deleting versions")
				matcher, err := semver.NewConstraint(versions)
				if err != nil {
					return err
				}
				deletions := []repo.BundleVersion{}
				for _, ver := range vers {
					if ok, _ := matcher.Validate(ver.Version); ok {
						fmt.Fprintf(w, "Version %s matches constraint %q\n", ver, versions)
						deletions = append(deletions, ver)
						index.DeleteVersion(bname, ver.Version.String())
						// If there are no more versions, remove the entire entry.
						if vers, ok := index.GetVersions(bname); ok && len(vers) == 0 {
							index.Delete(bname)
						}

					}
				}
				// We can skip writing the file if there is nothing to delete.
				if len(deletions) == 0 {
					return nil
				}
				if err := index.WriteFile(h.Repositories(), 0644); err != nil {
					return err
				}
				deleteBundleVersions(deletions, index, h, w)
				return nil
			}

			// If no version was specified, delete entire record
			if !index.Delete(bname) {
				fmt.Fprintf(w, "Bundle %q not found. Nothing deleted.", bname)
				return nil
			}
			if err := index.WriteFile(h.Repositories(), 0644); err != nil {
				return err
			}

			deleteBundleVersions(vers, index, h, w)
			return nil
		},
	}
	cmd.Flags().StringVarP(&versions, "version", "r", "", "A version or SemVer2 version range")

	return cmd
}

// deleteBundleVersions removes the given SHAs from bundle storage
//
// It warns, but does not fail, if a given SHA is not found.
func deleteBundleVersions(vers []repo.BundleVersion, index repo.Index, h home.Home, w io.Writer) {
	for _, ver := range vers {
		fpath := filepath.Join(h.Bundles(), ver.Digest)
		if err := os.Remove(fpath); err != nil {
			fmt.Fprintf(w, "WARNING: could not delete stake record %q", fpath)
		}
	}
}
