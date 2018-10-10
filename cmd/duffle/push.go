package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/duffle/home"
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

			b, err := loadBundle(push.bundleFile)
			if err != nil {
				return err
			}

			url, err := getBundleRepoURL(fmt.Sprintf("%s/%s:%s", push.repo, b.Name, b.Version), push.home)
			if err != nil {
				return err
			}

			body, err := os.Open(push.bundleFile)
			if err != nil {
				return err
			}

			req, err := http.NewRequest("POST", url.String(), body)
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			fmt.Fprintf(push.out, "Successfully pushed %s:%s to %s\n", b.Name, b.Version, push.repo)
			return nil
		},
	}

	cmd.Flags().StringVarP(&push.bundleFile, "file", "f", "", "bundle file to push")
	cmd.Flags().StringVarP(&push.repo, "repo", "", "https://hub.cnlabs.io", "repo to push to")

	return cmd
}
