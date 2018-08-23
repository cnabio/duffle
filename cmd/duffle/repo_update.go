package main

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/ohai"
	"github.com/deis/duffle/pkg/repo/installer"

	"github.com/spf13/cobra"
)

func newRepoUpdateCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update repositories",
		RunE: func(_ *cobra.Command, _ []string) error {
			return updateRepositories(w)
		},
	}
	return cmd
}

func updateRepositories(w io.Writer) error {
	start := time.Now()
	home := home.Home(homePath())
	repositories := findRepositories(home.Repositories())
	for _, repository := range repositories {
		i, err := installer.FindSource(filepath.Join(home.Repositories(), repository), home)
		if err != nil {
			return err
		}
		if err := installer.Update(i); err != nil {
			return err
		}
	}
	t := time.Now()
	ohai.Fohailn(w, "repositories updated!")
	for _, repository := range repositories {
		fmt.Fprintln(w, repository)
	}
	fmt.Fprintln(w, "")
	ohai.Fsuccessf(w, "repositories updated in %s\n", t.Sub(start).String())
	return nil
}
