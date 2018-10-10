package main

import (
	"errors"
	"fmt"
	"io"
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
	var identity string
	cmd := &cobra.Command{
		Use:   "sign FILE",
		Short: "clear-sign a bundle.json file",
		Long:  keySignDesc,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			h := home.Home(homePath())
			secring := h.SecretKeyRing()
			return signFile(args[0], secring, identity)
		},
	}
	cmd.Flags().StringVarP(&identity, "user", "u", "", "the user ID of the key to use. Format is either email address or 'NAME (COMMENT) <EMAIL>'")

	return cmd
}

func signFile(filepath, keyring, identity string) error {
	// Verify that file exists
	if fi, err := os.Stat(filepath); err != nil {
		return err
	} else if fi.IsDir() {
		return errors.New("cannot sign a directory")
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
	data, err := s.Clearsign(&bundle.Bundle{Name: "test"})
	fmt.Println(string(data))
	return err
}
