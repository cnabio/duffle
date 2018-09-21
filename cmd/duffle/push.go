package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/spf13/cobra"
)

type pushCmd struct {
	out  io.Writer
	src  string
	home home.Home
}

func newPushCmd(out io.Writer) *cobra.Command {
	const usage = `pushes a CNAB bundle to a repository`

	var push = &pushCmd{out: out}

	cmd := &cobra.Command{
		Use:   "push",
		Short: usage,
		Long:  usage,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				push.src = args[0]
			}
			push.home = home.Home(homePath())
			return push.run()
		},
	}

	return cmd
}

func (p *pushCmd) run() error {
	bundlePath := filepath.Join(p.src, "cnab", "bundle.json")
	b, err := loadBundle(bundlePath)
	if err != nil {
		return err
	}

	bundleParts := strings.Split(b.Name, "/")
	repo := bundleParts[0]
	name := bundleParts[1]

	// TODO - decide on the api here - bundles vs. repositories
	//
	// this is the case of a thin bundle, where only the bundle file is pushed
	url := fmt.Sprintf("https://%s/bundles/%s.json", repo, name)

	body, err := os.Open(bundlePath)
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
