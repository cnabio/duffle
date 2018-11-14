package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/repo"
	"github.com/deis/duffle/pkg/repo/remote/auth"
)

func newPushCmd(out io.Writer) *cobra.Command {
	const usage = `Pushes a CNAB bundle to a repository.`

	cmd := &cobra.Command{
		Use:   "push NAME",
		Short: "push a CNAB bundle to a repository",
		Long:  usage,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			home := home.Home(homePath())
			bundleName := args[0]

			ref, err := getReference(bundleName)
			if err != nil {
				return fmt.Errorf("could not parse reference for %s: %v", bundleName, err)
			}

			// read the bundle reference from repositories.json
			index, err := repo.LoadIndex(home.Repositories())
			if err != nil {
				return fmt.Errorf("cannot open %s: %v", home.Repositories(), err)
			}

			digest, err := index.Get(ref.Name(), ref.Tag())
			if err != nil {
				return err
			}

			body, err := os.Open(filepath.Join(home.Bundles(), digest))
			if err != nil {
				return err
			}
			defer body.Close()

			url, err := repoURLFromReference(ref)
			if err != nil {
				return err
			}

			loginCreds, err := auth.Load(filepath.Join(home.String(), "auth.json"))
			if err != nil {
				return err
			}

			creds, err := loginCreds.Get(url.String())
			if err != nil {
				log.Debug(err)
				return errors.New("could not retrieve authorization token. Are you logged in?")
			}

			req, err := http.NewRequest("POST", url.String(), body)
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", creds.Token))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				return fmt.Errorf("token for %s expired. Please run `duffle login %s` again to fetch a new auth token", url.Hostname(), url.Hostname())
			} else if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("request to %s responded with a non-200 status code: %d", url, resp.StatusCode)
			}

			fmt.Fprintf(out, "Successfully pushed %s\n", ref.String())
			return nil
		},
	}

	return cmd
}
