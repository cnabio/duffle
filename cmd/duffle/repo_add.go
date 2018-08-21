package main

import (
	"io"
	"time"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/ohai"
	"github.com/deis/duffle/pkg/repo/installer"

	"github.com/spf13/cobra"
)

// TODO
func newRepoAddCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <repo> [name]",
		Short: "add repositories",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 1 {
				name = args[1]
			}
			i, err := installer.New(args[0], name, "", home.Home(homePath()))
			if err != nil {
				return err
			}

			start := time.Now()
			if err := installer.Install(i); err != nil {
				return err
			}
			t := time.Now()
			ohai.Fsuccessf(w, "repo added in %s\n", t.Sub(start).String())
			return nil
		},
	}
	return cmd
}
