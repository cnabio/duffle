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

type credentialListCmd struct {
	out  io.Writer
	home home.Home
}

func newCredentialListCmd(w io.Writer) *cobra.Command {

	list := &credentialListCmd{out: w}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list credential sets",
		RunE: func(cmd *cobra.Command, args []string) error {
			list.home = home.Home(homePath())
			return list.run()
		},
	}
	return cmd
}

func (ls *credentialListCmd) run() error {
	credentialPath := ls.home.Credentials()
	creds := findCredentialSets(credentialPath)

	for name, _ := range creds {
		fmt.Fprintln(ls.out, name)
	}
	return nil
}

func findCredentialSets(dir string) map[string]string {
	creds := map[string]string{} // name: path

	verbosePrint("Traversing credentials directory (%s) for credential sets", dir)

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !f.IsDir() {
			verbosePrint("Loading credential set from %s", path)
			credSet, err := credentials.Load(path)
			if err != nil {
				verbosePrint("Unable to load credential set from %s:\n%s", path, err)
				return nil
			}

			verbosePrint("Successfully loaded credential set %s from %s", credSet.Name, path)
			creds[credSet.Name] = path
		}
		return nil
	})

	return creds
}
