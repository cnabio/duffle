package main

import (
	"fmt"
	"io"

	builder2 "github.com/deislabs/duffle/pkg/imagestore/builder"

	"github.com/pkg/errors"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/loader"
	"github.com/deislabs/duffle/pkg/packager"

	"github.com/spf13/cobra"
)

const exportDesc = `
Packages a bundle, and by default any images referenced by the bundle, within
a single gzipped tarfile.

If neither --oci-layout nor --thin is specified, all images (incuding invocation
images) referenced by the bundle metadata are saved as tar files in the
artifacts/ directory along with an artifacts.json file which describes the
contents of the artifacts/ directory.

If --oci-layout is specified, all images (incuding invocation images) referenced
by the bundle metadata are saved as an OCI image layout in the artifacts/layout/
directory.

If --thin specified, only the bundle manifest is exported.

By default, this command will use the name and version information of
the bundle to create a compressed archive file called
<name>-<version>.tgz in the current directory. This destination can be
updated by specifying a file path to save the compressed bundle to using
the --output-file flag.

Pass in a path to a bundle file instead of a bundle in local storage by
using the --bundle-is-file flag like below:
$ duffle export [PATH] --bundle-is-file
`

type exportCmd struct {
	bundle       string
	dest         string
	home         home.Home
	out          io.Writer
	thin         bool
	verbose      bool
	bundleIsFile bool
	ociLayout    bool
}

func newExportCmd(w io.Writer) *cobra.Command {
	export := &exportCmd{out: w}

	cmd := &cobra.Command{
		Use:   "export [BUNDLE]",
		Short: "package CNAB bundle in gzipped tar file",
		Long:  exportDesc,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			export.home = home.Home(homePath())
			export.bundle = args[0]
			if export.thin && export.ociLayout {
				return errors.New("--thin and --oci-layout must not both be specified")
			}

			return export.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&export.dest, "output-file", "o", "", "Save exported bundle to file path")
	f.BoolVarP(&export.bundleIsFile, "bundle-is-file", "f", false, "Indicates that the bundle source is a file path")
	f.BoolVarP(&export.ociLayout, "oci-layout", "l", false, "Export images as an OCI image layout")
	f.BoolVarP(&export.thin, "thin", "t", false, "Export only the bundle manifest")
	f.BoolVarP(&export.verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func (ex *exportCmd) run() error {
	bundlefile, l, err := ex.setup()
	if err != nil {
		return err
	}
	if err := ex.Export(bundlefile, l); err != nil {
		return err
	}

	return nil
}

func (ex *exportCmd) Export(bundlefile string, l loader.BundleLoader) error {
	builder, err := builder2.NewBuilder(ex.thin, ex.ociLayout)
	if err != nil {
		return err
	}

	exp, err := packager.NewExporter(bundlefile, ex.dest, ex.home.Logs(), l, builder)
	if err != nil {
		return fmt.Errorf("Unable to set up exporter: %s", err)
	}
	if err := exp.Export(); err != nil {
		return err
	}
	if ex.verbose {
		fmt.Fprintf(ex.out, "Export logs: %s\n", exp.Logs())
	}
	return nil
}

func (ex *exportCmd) setup() (string, loader.BundleLoader, error) {
	l := loader.New()
	bundlefile, err := resolveBundleFilePath(ex.bundle, ex.home.String(), ex.bundleIsFile)
	if err != nil {
		return "", l, err
	}

	return bundlefile, l, nil
}

func resolveBundleFilePath(bun, homePath string, bundleIsFile bool) (string, error) {

	if bundleIsFile {
		return bun, nil
	}

	bundlefile, err := getBundleFilepath(bun, homePath)
	if err != nil {
		return "", err
	}
	return bundlefile, err
}
