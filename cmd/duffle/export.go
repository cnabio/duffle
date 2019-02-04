package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/scothis/ruffle/pkg/duffle/home"
	"github.com/scothis/ruffle/pkg/loader"
	"github.com/scothis/ruffle/pkg/packager"

	"github.com/spf13/cobra"
)

const exportDesc = `
Packages a bundle, invocation images, and all referenced images within a single
gzipped tarfile.

All images specified in the bundle metadata are saved as tar files in the artifacts/
directory along with an artifacts.json file which describes the contents of artifacts/.

By default, this command will use the name and version information of the bundle to create
a compressed archive file called <name>-<version>.tgz in the current directory. This
destination can be updated by specifying a file path to save the compressed bundle to using
the --output-file flag.

If you want to export only the bundle manifest without the invocation images and referenced 
images, use the --thin flag.
`

type exportCmd struct {
	bundleRef string
	dest      string
	home      home.Home
	out       io.Writer
	thin      bool
	verbose   bool
	insecure  bool
}

func newExportCmd(w io.Writer) *cobra.Command {
	export := &exportCmd{out: w}

	cmd := &cobra.Command{
		Use:   "export [BUNDLE]",
		Short: "package CNAB bundle in gzipped tar file",
		Long:  exportDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("this command requires an argument: a bundle")
			}
			export.home = home.Home(homePath())
			export.bundleRef = args[0]

			return export.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&export.dest, "output-file", "o", "", "Save exported bundle to file path")
	f.BoolVarP(&export.thin, "thin", "t", false, "Export only the bundle manifest")
	f.BoolVarP(&export.verbose, "verbose", "v", false, "Verbose output")
	f.BoolVarP(&export.insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")

	return cmd
}

func (ex *exportCmd) run() error {
	source, l, err := ex.setup()
	if err != nil {
		return err
	}
	if ex.Export(source, l); err != nil {
		return err
	}

	return nil
}

func (ex *exportCmd) setup() (string, loader.Loader, error) {
	source, err := getBundleFilepath(ex.bundleRef, ex.home.String(), ex.insecure)
	if err != nil {
		return "", nil, err
	}

	l, err := getLoader(ex.home.String(), ex.insecure)
	if err != nil {
		return "", nil, err
	}

	return source, l, nil
}

func (ex *exportCmd) Export(source string, l loader.Loader) error {
	exp, err := packager.NewExporter(source, ex.dest, ex.home.Logs(), l, ex.thin, ex.insecure)
	if err != nil {
		return fmt.Errorf("Unable to set up exporter: %s", err)
	}
	if err := exp.Export(); err != nil {
		return err
	}
	if ex.verbose {
		fmt.Fprintf(ex.out, "Export logs: %s\n", exp.Logs)
	}
	return nil
}
