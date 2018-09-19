package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/duffle/home"
)

const credentialGenerateHelp = `Generate credentials from a CNAB bundle

This reads a bundle.json file's credential requirements and generates a stub credentialset.
The given name becomes the name of the credentialset.

If a bundle is given, the bundle may be fetched (unless there is a cached copy),
and will then be examined. If the '-f' flag is specified, though, it will read the
bundle.json supplied.

The generated credentials will all be initialed to stub values, and should be edited
to reflect the true values.

The newly created credential set will be added to the credentialsets, though users
will still need to edit that file to set the appropriate values.
`

func newCredentialGenerateCmd(out io.Writer) *cobra.Command {
	bundleFile := ""
	cmd := &cobra.Command{
		Use:     "generate NAME [BUNDLE]",
		Aliases: []string{"gen"},
		Short:   "generate a credentialset from a bundle",
		Long:    credentialGenerateHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			bf, err := bundleFileOrArg2(args, bundleFile, out)
			if err != nil {
				return err
			}
			csName := args[0]

			bun, err := loadBundle(bf)
			if err != nil {
				return err
			}

			creds := genCredentialSet(csName, bun.Credentials)
			//data, err := json.MarshalIndent(creds, "", "  ")
			data, err := yaml.Marshal(creds)
			if err != nil {
				return err
			}

			fmt.Printf("%v", string(data))

			dest := filepath.Join(home.Home(homePath()).Credentials(), csName+".yaml")
			return ioutil.WriteFile(dest, data, 0600)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&bundleFile, "file", "f", "", "path to bundle.json")
	return cmd
}

func genCredentialSet(name string, creds map[string]bundle.CredentialLocation) credentials.CredentialSet {
	cs := credentials.CredentialSet{
		Name: name,
	}
	cs.Credentials = []credentials.CredentialStrategy{}

	for name, loc := range creds {
		c := credentials.CredentialStrategy{
			Name:        name,
			Source:      credentials.Source{Value: "EMPTY"},
			Destination: credentials.Destination{},
		}
		if loc.EnvironmentVariable != "" {
			c.Destination.EnvVar = loc.EnvironmentVariable
		}
		if loc.Path != "" {
			c.Destination.Path = loc.Path
		}
		cs.Credentials = append(cs.Credentials, c)
	}

	return cs
}
