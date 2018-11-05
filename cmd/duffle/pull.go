package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/deis/duffle/pkg/duffle/home"
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
			home := home.Home(homePath())
			url, err := getBundleRepoURL(args[0], home)
			if err != nil {
				return err
			}
			b, err := loadBundle(url.String(), insecure)
			bundleFilepath := filepath.Join(home.Cache(), fmt.Sprintf("%s-%s.json", strings.Replace(b.Name, "/", "-", -1), b.Version))
			if err := b.WriteFile(bundleFilepath, 0644); err != nil {
				return err
			}
			fmt.Fprintln(w, bundleFilepath)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")

	return cmd
}
