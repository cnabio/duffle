package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/reference"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/docker/cnab-to-oci/remotes"
	"github.com/spf13/cobra"
)

type pullCmd struct {
	output             string
	targetRef          string
	insecureRegistries []string
	home               home.Home
}

func newPullCmd(w io.Writer) *cobra.Command {
	const usage = `Pulls a bundle from an OCI repository`
	const pullDesc = `
Pulls a CNAB bundle from an OCI repository.
The only argument for this command is the repository where
the bundle can be found, and by default, this command pulls the
bundle and stores it in the local bundle store.

If the --output flag is passed, the bundle file will be saved in
that file, and its reference will not be recorded in the local store.

Insecure registries can be passed through the --insecure-registries flags.

Examples:
$ duffle pull registry/username/bundle:tag
$ duffle pull --output path-for-bundle.json registry/username/bundle:tag
`
	var pull pullCmd
	cmd := &cobra.Command{
		Use:   "pull [TARGET REPOSITORY] [options]",
		Short: usage,
		Long:  pullDesc,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pull.targetRef = args[0]
			pull.home = home.Home(homePath())
			return pull.run()
		},
	}

	cmd.Flags().StringVarP(&pull.output, "output", "o", "", "Output file")
	cmd.Flags().StringSliceVar(&pull.insecureRegistries, "insecure-registries", nil, "Use plain HTTP for those registries")
	return cmd
}

func (p *pullCmd) run() error {
	ref, err := reference.ParseNormalizedNamed(p.targetRef)
	if err != nil {
		return err
	}
	b, err := remotes.Pull(context.Background(), ref, createResolver(p.insecureRegistries))
	if err != nil {
		return err
	}

	return p.writeBundle(b)
}

func (p *pullCmd) writeBundle(bf *bundle.Bundle) error {
	data, digest, err := marshalBundle(bf)
	if err != nil {
		return fmt.Errorf("cannot marshal bundle: %v", err)
	}

	if p.output != "" {
		if err := ioutil.WriteFile(p.output, data, 0644); err != nil {
			return fmt.Errorf("cannot write bundle to %s: %v", p.output, err)
		}
		return nil
	}

	if err := ioutil.WriteFile(filepath.Join(p.home.Bundles(), digest), data, 0644); err != nil {
		return fmt.Errorf("cannot store bundle : %v", err)

	}

	return recordBundleReference(p.home, bf.Name, bf.Version, digest)

}
