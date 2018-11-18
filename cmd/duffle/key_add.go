package main

import (
	"io"
	"os"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/signature"

	"github.com/spf13/cobra"
)

const keyAddDesc = `Add a key or keys to the keyring.

Add keys to either the public (default) or secret (-s) keyring. By default, this
expects binary keys (in the form generated by 'duffle key export'), but with the
'--armored'/'-a' flag this can take ASCII armored keys as well.

Keys added to the secret keyring must contain private key material. Keys added to the
public keyring should be public keys, but private keys will be accepted (though the
private key material may be removed).
`

func newKeyAddCmd(w io.Writer) *cobra.Command {
	var secret bool
	var armored bool
	cmd := &cobra.Command{
		Use:   "add FILE",
		Short: "add one or more keys to the keyring",
		Long:  keyAddDesc,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			h := home.Home(homePath())
			var ring string
			if secret {
				// If secret, add one to the secret keyring as well as the public
				// eyring. When added to the public keyring, private material will
				// be stripped by `SavePublic`. By doing this, we make sure that
				// any key added to the secret ring can be used to verify a
				// bundle.
				ring = h.SecretKeyRing()
				if err := addKeys(args[0], ring, secret, armored); err != nil {
					return err
				}
			}
			ring = h.PublicKeyRing()
			return addKeys(args[0], ring, false, armored)
		},
	}
	cmd.Flags().BoolVarP(&secret, "secret", "s", false, "Add a secret (private) key")
	cmd.Flags().BoolVarP(&armored, "armored", "a", false, "Load an ASCII armored key")
	return cmd
}

func addKeys(file, ring string, private, armored bool) error {
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
	if err := kring.Add(reader, armored); err != nil {
		return err
	}
	if private {
		return kring.SavePrivate(ring, true)
	}
	return kring.SavePublic(ring, true)
}
