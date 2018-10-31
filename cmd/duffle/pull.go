package main

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/crypto/digest"
	"github.com/deis/duffle/pkg/duffle/home"
)

func newPullCmd(w io.Writer) *cobra.Command {
	const usage = `Pulls a CNAB bundle into the cache without installing it.

Example:
	$ duffle pull duffle/example:0.1.0
`

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "pull a CNAB bundle from a repository",
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := pullBundle(args[0], home.Home(homePath()))
			if err != nil {
				return err
			}
			fmt.Fprintln(w, path)
			return nil
		},
	}

	return cmd
}

func pullBundle(bundleName string, home home.Home) (string, error) {
	ref, err := getReference(bundleName)
	if err != nil {
		return "", err
	}

	url, err := getBundleRepoURL(bundleName)
	if err != nil {
		return "", err
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request to %s responded with a non-200 status code: %d", url, resp.StatusCode)
	}

	buf, digest, err := digest.OfReader(resp.Body)
	if err != nil {
		return "", err
	}

	bundle, err := bundle.ParseReader(buf)
	if err != nil {
		return "", err
	}
	bundleFilepath := filepath.Join(home.Bundles(), digest)
	if err := bundle.WriteFile(bundleFilepath, 0644); err != nil {
		return "", fmt.Errorf("failed to write bundle: %v", err)
	}

	return bundleFilepath, recordBundleReference(home, ref.Name(), ref.Tag(), digest)
}
