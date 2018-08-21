package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/spf13/cobra"
)

func newRepoListCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := home.Home(homePath()).Repositories()
			rigs := findRepositories(repoPath)

			for _, rig := range rigs {
				fmt.Fprintln(w, rig)
			}
			return nil
		},
	}
	return cmd
}

func findRepositories(dir string) []string {
	var rigs []string
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() && f.Name() == "bundles" {
			rigName := strings.TrimPrefix(filepath.Dir(path), dir+string(os.PathSeparator))
			rigs = append(rigs, rigName)
		}
		return nil
	})
	return rigs
}
