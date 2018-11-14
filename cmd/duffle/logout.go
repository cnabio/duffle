package main

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/bacongobbler/browser"
	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/repo/remote/auth"
)

type logoutCmd struct {
	out  io.Writer
	home home.Home
}

func newLogoutCmd(out io.Writer) *cobra.Command {
	logout := logoutCmd{out: out}
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "log out",
		RunE: func(cmd *cobra.Command, args []string) error {
			logout.home = home.Home(homePath())
			return logout.run(args[0])
		},
	}

	return cmd
}

func (l *logoutCmd) run(hostname string) error {
	url := &url.URL{
		Scheme: "https",
		Host:   hostname,
		Path:   "/.auth/logout",
	}
	if err := browser.Open(url.String()); err != nil {
		fmt.Printf("please open %s in your browser to log out.\n", url.String())
	}

	loginCreds, err := auth.Load(filepath.Join(l.home.String(), "auth.json"))
	if err != nil {
		return err
	}
	loginCreds.Remove(url.Hostname())
	return loginCreds.WriteFile(filepath.Join(l.home.String(), "auth.json"), 0644)
}
