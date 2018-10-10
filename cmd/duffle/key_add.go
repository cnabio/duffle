package main

import (
	"io"
	"os"

	"github.com/deis/duffle/pkg/signature"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/spf13/cobra"
)

const keyAddDesc = `Add a key or keys to the keyring.

Add keys to either the public (default) or secret (-s) keyring. The file must be an ASCII-armored
key or keyring.

Keys added to the secret keyring must contain private key material. Keys added to the
public keyring should be public keys, but private keys will be accepted (though the
private key material may be removed).
`

func newKeyAddCmd(w io.Writer) *cobra.Command {
	var secret bool
	cmd := &cobra.Command{
		Use:   "add FILE",
		Short: "add one or more keys to the keyring",
		Long:  keyAddDesc,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			h := home.Home(homePath())
			var ring string
			if secret {
				ring = h.SecretKeyRing()
			} else {
				ring = h.PublicKeyRing()
			}

			return addKeys(args[0], ring, secret)
		},
	}
	cmd.Flags().BoolVarP(&secret, "secret", "s", false, "add a secret (private) key")
	return cmd
}

func addKeys(file, ring string, private bool) error {
	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer reader.Close()
	kring, err := signature.LoadKeyRing(ring)
	if err != nil {
		return err
	}
	kring.PassphraseFetcher = passwordFetcher
	if err := kring.Add(reader); err != nil {
		return err
	}
	if private {
		return kring.SavePrivate(ring, true)
	}
	return kring.SavePublic(ring, true)
}
