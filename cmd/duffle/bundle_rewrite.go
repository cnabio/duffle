package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/bundle/rewrite"
	"github.com/deis/duffle/pkg/crypto/digest"
	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/signature"

	"github.com/spf13/cobra"
)

type bundleRewriteCmd struct {
	out            io.Writer
	home           home.Home
	bundleFile     string
	skipValidation bool
	signer         string
}

func newBundleRewriteCmd(w io.Writer) *cobra.Command {

	const usage = `Rewrite repository references in an existing CNAB bundle

	The rewrite command takes a local CNAB bundle.json file and rewrites the invocation
	image and any images defined in the images array of the bundle.json file. The command will also
	update file references for the images. Images will also be retagged. Docker images must exist locally.
	
	The repository argument will be used to rewrite the image.
	
	Example:
			$ duffle rewrite BUNDLENAME deis/example
	
			Running this command against a bundle.json with an invocation image defined as cnab/hellohelm:latest
			will result in the bundle.json being rewritten to deis/example/hellohelm:latest
	
	After rewriting a bundle, you will need to push the newly tagged images.
	`

	rw := bundleRewriteCmd{out: w}

	cmd := &cobra.Command{
		Use:   "rewrite BUNDLE NEW_REGISTRY",
		Short: "rewrite repository references in bundle",
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			rw.home = home.Home(homePath())
			w := cmd.OutOrStdout()
			bundleFile, newRegistry, err := validateRewriteArgs(args, rw.bundleFile, w, rw.skipValidation)
			if err != nil {
				return err
			}
			bun, err := loadBundle(bundleFile, rw.skipValidation)
			if err != nil {
				return err
			}
			rewriter, err := rewrite.New(w)
			if err != nil {
				return err
			}
			ctx := context.Background()
			err = rewriter.Rewrite(ctx, bun, newRegistry)
			if err != nil {
				fmt.Fprintf(w, "error creating bundle rewriter: %s\n", err)
				return err
			}
			digest, err := rw.rewriteBundle(bun)
			if err != nil {
				fmt.Fprintf(w, "error rewriting bundle: %s\n", err)
				return err
			}
			// record the new bundle in repositories.json
			if err := recordBundleReference(rw.home, bun.Name, bun.Version, digest); err != nil {
				return fmt.Errorf("could not record bundle: %v", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&rw.signer, "user", "u", "", "the user ID of the signing key to use. Format is either email address or 'NAME (COMMENT) <EMAIL>'")
	cmd.Flags().BoolVarP(&rw.skipValidation, "insecure", "k", false, "Do not verify the bundle (INSECURE)")
	cmd.Flags().StringVarP(&rw.bundleFile, "file", "f", "", "path to bundle file to sign")

	return cmd
}

func validateRewriteArgs(args []string, bundleFile string, w io.Writer, insecure bool) (string, string, error) {
	switch {
	case len(args) < 1:
		return "", "", errors.New("this command requires at least one argument: REGISTRY (the new registry name). It also requires a BUNDLE (CNAB bundle name) or file (using -f)\nValid inputs:\n\t$ duffle bundle rewrite BUNDLE REGISTRY\n\t$ duffle bundle rewrite REGISTRY -f path-to-bundle.json")
	case len(args) == 2 && bundleFile != "":
		return "", "", errors.New("please use either -f or specify a BUNDLE, but not both")
	case len(args) < 2 && bundleFile == "":
		return "", "", errors.New("required arguments are BUNDLE (CNAB bundle name) or file and REGISTRY (Docker Registry)")
	case len(args) == 2:
		bun, err := loadOrPullBundle(args[0], insecure)
		return bun, args[1], err
	}
	return bundleFile, args[0], nil
}

func (b bundleRewriteCmd) rewriteBundle(bf *bundle.Bundle) (string, error) {
	kr, err := signature.LoadKeyRing(b.home.SecretKeyRing())
	if err != nil {
		return "", fmt.Errorf("cannot load keyring: %s", err)
	}

	if kr.Len() == 0 {
		return "", errors.New("no signing keys are present in the keyring")
	}

	// Default to the first key in the ring unless the user specifies otherwise.
	key := kr.Keys()[0]
	if b.signer != "" {
		key, err = kr.Key(b.signer)
		if err != nil {
			return "", err
		}
	}

	sign := signature.NewSigner(key)
	data, err := sign.Clearsign(bf)
	data = append(data, '\n')
	if err != nil {
		return "", fmt.Errorf("cannot sign bundle: %s", err)
	}

	digest, err := digest.OfBuffer(data)
	if err != nil {
		return "", fmt.Errorf("cannot compute digest from bundle: %v", err)
	}

	return digest, ioutil.WriteFile(filepath.Join(b.home.Bundles(), digest), data, 0644)
}
