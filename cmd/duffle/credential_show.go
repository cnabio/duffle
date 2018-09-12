package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/duffle/home"
)

const credentialShowDesc = `
This command will fetch the credential set with the given name and prints the contents of the file.
`

type credentialShowCmd struct {
	name string
	home home.Home
	out  io.Writer
}

func newCredentialShowCmd(w io.Writer) *cobra.Command {
	show := &credentialShowCmd{out: w}

	cmd := &cobra.Command{
		Use:   "show [NAME]",
		Short: "show credential set",
		RunE: func(cmd *cobra.Command, args []string) error {
			show.home = home.Home(homePath())
			show.name = args[0]
			return show.run()
		},
	}
	return cmd
}

func (sh *credentialShowCmd) run() error {
	cs, err := findCredentialSet(sh.home.Credentials(), sh.name)
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(cs.Name)
	if err != nil {
		return err
	}
	fmt.Printf("name: %s", string(b))
	b, err = yaml.Marshal(cs.Credentials)
	if err != nil {
		return err
	}
	fmt.Printf("credentials:\n%s", string(b))
	return nil
}

func findCredentialSet(dir, name string) (*credentials.CredentialSet, error) {
	var cs *credentials.CredentialSet
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !f.IsDir() {
			credSet, err := credentials.Load(path)
			if err != nil {
				return err
			}
			cs = credSet
			return nil
		}
		return nil
	})
	return cs, nil
}
