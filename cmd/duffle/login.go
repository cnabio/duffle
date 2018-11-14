package main

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/bacongobbler/browser"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/repo/remote/auth"
)

type loginCmd struct {
	out  io.Writer
	home home.Home
}

func newLoginCmd(out io.Writer) *cobra.Command {
	login := loginCmd{out: out}
	cmd := &cobra.Command{
		Use:   "login",
		Short: "log in",
		RunE: func(cmd *cobra.Command, args []string) error {
			login.home = home.Home(homePath())
			return login.run(args[0])
		},
	}

	return cmd
}

func (l *loginCmd) run(hostname string) error {
	url := &url.URL{
		Scheme: "https",
		Host:   hostname,
		Path:   "/.auth/me",
	}
	fmt.Println("In order to log in, we will attempt to open a webpage in your browser which will prompt for your information. Once logged in, please enter your auth token below.")

	if err := browser.Open(url.String()); err != nil {
		fmt.Printf("Please open %s in your browser to log in.\n", url.String())
	}

	prompt := &survey.Password{
		Message: "Please enter your auth token",
	}

	var token string
	if err := survey.AskOne(prompt, &token, nil); err != nil {
		return err
	}

	loginCreds, err := auth.Load(filepath.Join(l.home.String(), "auth.json"))
	if err != nil {
		return err
	}
	loginCreds.Add(url.Hostname(), auth.RepositoryCredentials{Token: token})
	return loginCreds.WriteFile(filepath.Join(l.home.String(), "auth.json"), 0644)
}
