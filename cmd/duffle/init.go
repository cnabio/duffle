package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/ohai"
	"github.com/deis/duffle/pkg/signature"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/spf13/cobra"
)

const (
	initDesc = `
Initializes duffle with configuration required to start installing CNAB bundles.

This command will create a subdirectory in your home directory, and use that directory for storing
configuration, preferences, and persistent data. Duffle uses OpenPGP-style keys for signing and
verification. If you do not provide a keyring to import, the init phase will generate a keyring for
you, and create a signing key.
`
)

type initCmd struct {
	dryRun   bool
	keyFile  string
	username string
	w        io.Writer
}

func newInitCmd(w io.Writer) *cobra.Command {
	i := &initCmd{w: w}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "sets up local environment to work with duffle",
		Long:  initDesc,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.run()
		},
	}

	f := cmd.Flags()
	f.BoolVar(&i.dryRun, "dry-run", false, "go through all the steps without actually installing anything")
	f.StringVarP(&i.keyFile, "signing-key", "k", "", "Armored OpenPGP key to be used for signing. If not specified, one will be generated for you")
	f.StringVarP(&i.username, "user", "u", os.ExpandEnv("$USER <$USER@localhost>"), "User identity for the OpenPGP key. It is best to set this to your email address.")

	return cmd
}

func (i *initCmd) run() error {
	home := home.Home(homePath())
	dirs := []string{
		home.String(),
		home.Logs(),
		home.Plugins(),
		home.Repositories(),
		home.Claims(),
		home.Credentials(),
	}

	if err := i.ensureDirectories(dirs); err != nil {
		return err
	}
	if err := i.ensureRepositories(); err != nil {
		return err
	}
	if _, err := i.loadOrCreateSecretKeyRing(home.SecretKeyRing()); err != nil {
		return err
	}
	_, err := i.loadOrCreatePublicKeyRing(home.PublicKeyRing())
	return err
}

// ensureRepositories checks to see if the default repositories exists.
//
// If the repo does not exist, this function will create it.
func (i *initCmd) ensureRepositories() error {
	ohai.Fohailn(i.w, "Installing default repositories...")

	// TODO: add repos here
	addArgs := []string{"ssh://git@github.com/deis/bundles.git"}

	repoCmd, _, err := rootCmd.Find([]string{"repo", "add"})
	if err != nil {
		return err
	}
	if i.dryRun {
		return nil
	}
	return repoCmd.RunE(repoCmd, addArgs)
}

func (i *initCmd) ensureDirectories(dirs []string) error {
	fmt.Fprintln(i.w, "The following new directories will be created:")
	fmt.Fprintln(i.w, strings.Join(dirs, "\n"))
	for _, dir := range dirs {
		if fi, err := os.Stat(dir); err != nil {
			if !i.dryRun {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("Could not create %s: %s", dir, err)
				}
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", dir)
		}
	}
	return nil
}

// This loads a keyring from disk. If no keyring already exists, this will create a new
// keyring, add a new default identity, and then write that keyring to disk.
//
// Regardless of the path, a *signature.KeyRing will be returned.
func (i *initCmd) loadOrCreateSecretKeyRing(dest string) (*signature.KeyRing, error) {
	if _, err := os.Stat(dest); err == nil {
		// Since this is non-mutating, we can do this in a dry-run.
		return signature.LoadKeyRing(dest)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	fmt.Fprintf(i.w, "==> Generating a new secret keyring at %s\n", dest)

	// We could probably move the dry-run to just before the `ring.Save`. Not sure
	// what that accomplishes, though.
	if i.dryRun {
		return &signature.KeyRing{}, nil
	}

	ring := signature.CreateKeyRing(passwordFetcher)
	if i.keyFile != "" {
		key, err := os.Open(i.keyFile)
		if err != nil {
			return ring, err
		}
		err = ring.Add(key)
		key.Close()
		if err != nil {
			return ring, err
		}
		for _, k := range ring.PrivateKeys() {
			fmt.Fprintf(i.w, "==> Importing %q\n", k.UserID())
		}
	} else {
		var user signature.UserID
		if i.username != "" {
			var err error
			user, err = signature.ParseUserID(i.username)
			if err != nil {
				return ring, err
			}
		} else {
			user = defaultUserID()

		}
		// Generate the key
		fmt.Fprintf(i.w, "==> Generating a new signing key with ID %s\n", user.String())
		k, err := signature.CreateKey(user)
		if err != nil {
			return ring, err
		}
		ring.AddKey(k)
	}
	return ring, ring.Save(dest, false)
}

func (i *initCmd) loadOrCreatePublicKeyRing(dest string) (*signature.KeyRing, error) {
	if _, err := os.Stat(dest); err == nil {
		// Since this is non-mutating, we can do this in a dry-run.
		return signature.LoadKeyRing(dest)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	fmt.Fprintf(i.w, "==> Generating a new public keyring at %s\n", dest)

	// We could probably move the dry-run to just before the `ring.Save`. Not sure
	// what that accomplishes, though.
	if i.dryRun {
		return &signature.KeyRing{}, nil
	}
	ring := signature.CreateKeyRing(passwordFetcher)
	return ring, ring.Save(dest, false)
}

func passwordFetcher(prompt string) ([]byte, error) {
	fmt.Printf("Passphrase for key %q >  ", prompt)
	pp, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return pp, err
}
