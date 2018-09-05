package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/duffle/home"
)

func newCredentialListCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list credential sets",
		RunE: func(cmd *cobra.Command, args []string) error {
			credentialPath := home.Home(homePath()).Credentials()
			creds := findCredentialSets(credentialPath)

			for _, cred := range creds {
				fmt.Fprintln(w, cred)
			}
			return nil
		},
	}
	return cmd
}

func findCredentialSets(dir string) []string {
	var creds []string
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !f.IsDir() {
			credSet, err := credentials.Load(path)
			if err != nil {
				return err //TODO: collect errs in debug output
			}
			creds = append(creds, credSet.Name)
		}
		return nil
	})
	return creds
}
