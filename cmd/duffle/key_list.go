package main

import (
	"fmt"
	"io"

	"github.com/deis/duffle/pkg/signature"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/duffle/home"
)

const keyListDesc = `List key IDs for either the public or private keychain.

By default, this lists all of the IDs in the private keychain (the ones you use to sign or attest).
`

func newKeyListCmd(w io.Writer) *cobra.Command {
	var public bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list key IDs",
		Long:  keyListDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			h := home.Home(homePath())
			ring := h.SecretKeyRing()
			if public {
				ring = h.PublicKeyRing()
			}
			return listKeys(cmd.OutOrStdout(), ring)
		},
	}
	cmd.Flags().BoolVarP(&public, "public", "p", false, "show public key IDs instead of private key IDs")

	return cmd
}

func listKeys(out io.Writer, ring string) error {
	kr, err := signature.LoadKeyRing(ring)
	if err != nil {
		return err
	}

	for _, k := range kr.Keys() {
		var name, fingerprint string
		id, err := k.UserID()
		if err != nil {
			name = "[anonymous key]"
		} else {
			name = id.String()
		}
		fingerprint = k.Fingerprint()
		fmt.Printf("%s\t%q\n", name, fingerprint)
	}

	return nil
}
