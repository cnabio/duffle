package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/signature"
)

const keySignDesc = `Clear-sign a given bundle.json file.

This remarshals the bundle.json into canonical form, and then clear-signs the JSON.
The output is written to STDOUT.

If no key name is supplied, this uses the first signing key in the secret keyring.
`

func newKeySignCmd(w io.Writer) *cobra.Command {
	var (
		identity   string
		bundleFile string
		outfile    string
		noValidate bool
	)

	cmd := &cobra.Command{
		Use:   "sign",
		Short: "clear-sign a bundle.json file",
		Long:  keySignDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			h := home.Home(homePath())
			secring := h.SecretKeyRing()
			return signFile(bundleFile, secring, identity, outfile, noValidate)
		},
	}
	cmd.Flags().StringVarP(&identity, "user", "u", "", "the user ID of the key to use. Format is either email address or 'NAME (COMMENT) <EMAIL>'")
	cmd.Flags().StringVarP(&bundleFile, "file", "f", "", "path to bundle file to sign")
	cmd.Flags().StringVarP(&outfile, "output-file", "o", "bundle.cnab", "the name of the output file")
	cmd.Flags().BoolVar(&noValidate, "no-validate", false, "do not validate the JSON before marshaling it.")

	return cmd
}

func signFile(filepath, keyring, identity, outfile string, skipValidation bool) error {
	// Verify that file exists
	if fi, err := os.Stat(filepath); err != nil {
		return fmt.Errorf("cannot find bundle file to sign: %v", err)
	} else if fi.IsDir() {
		return errors.New("cannot sign a directory")
	}

	bdata, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	b, err := bundle.Unmarshal(bdata)
	if err != nil {
		return err
	}

	if !skipValidation {
		if err := b.Validate(); err != nil {
			return err
		}
	}

	// Load keyring
	kr, err := signature.LoadKeyRing(keyring)
	if err != nil {
		return err
	}
	// Find identity
	var k *signature.Key
	if identity != "" {
		k, err = kr.Key(identity)
		if err != nil {
			return err
		}
	} else {
		all := kr.PrivateKeys()
		if len(all) == 0 {
			return errors.New("no private keys found")
		}
		k = all[0]
	}

	// Sign the file
	s := signature.NewSigner(k)
	data, err := s.Clearsign(b)
	data = append(data, '\n')
	ioutil.WriteFile(outfile, data, 0644)
	return err
}
