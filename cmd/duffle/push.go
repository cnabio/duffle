package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/spf13/cobra"
)

type pushCmd struct {
	out        io.Writer
	bundleFile string
	repo       string
	home       home.Home
}

func newPushCmd(out io.Writer) *cobra.Command {
	const usage = `pushes a CNAB bundle to a repository`

	var push = &pushCmd{out: out}

	cmd := &cobra.Command{
		Use:   "push",
		Short: usage,
		Long:  usage,
		RunE: func(_ *cobra.Command, args []string) error {
			push.home = home.Home(homePath())
			return push.run()
		},
	}

	cmd.Flags().StringVarP(&push.bundleFile, "file", "f", "", "bundle file to push")
	cmd.Flags().StringVarP(&push.repo, "repo", "", "", "repo to push to")

	return cmd
}

func (p *pushCmd) run() error {
	b, err := loadBundle(p.bundleFile)
	if err != nil {
		return err
	}

	// TODO - decide on the api here - bundles vs. repositories
	//
	// this is the case of a thin bundle, where only the bundle file is pushed
	url := fmt.Sprintf("https://%s/bundles/%s.json", p.repo, b.Name)

	body, err := os.Open(p.bundleFile)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
