package main

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/ohai"
	"github.com/spf13/cobra"
)

func newRepoRemoveCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <repository...>",
		Short: "remove repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			repoPath := home.Home(homePath()).Repositories()
			repositories := findRepositories(repoPath)
			foundRepositories := map[string]bool{}
			for _, arg := range args {
				foundRepositories[arg] = false
			}
			for _, repo := range repositories {
				for _, arg := range args {
					if repo == arg {
						foundRepositories[repo] = true
						if err := os.RemoveAll(filepath.Join(repoPath, repo)); err != nil {
							return err
						}
					}
				}
			}
			t := time.Now()
			for repo, found := range foundRepositories {
				if !found {
					ohai.Fwarningf(w, "repository '%s' was not found in the repository list\n", repo)
				}
			}
			ohai.Fsuccessf(w, "repositories uninstalled in %s\n", t.Sub(start).String())
			return nil
		},
	}
	return cmd
}
