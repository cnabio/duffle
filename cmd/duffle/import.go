package main

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/spf13/cobra"

	"github.com/cnabio/duffle/pkg/duffle/home"
	"github.com/cnabio/duffle/pkg/packager"
)

const importDesc = `
Unpacks a bundle from a gzipped tar file on local file system
`

type importCmd struct {
	source  string
	dest    string
	out     io.Writer
	home    home.Home
	verbose bool
}

func newImportCmd(w io.Writer) *cobra.Command {
	importc := &importCmd{
		out:  w,
		home: home.Home(homePath()),
	}

	cmd := &cobra.Command{
		Use:   "import [PATH]",
		Short: "unpack CNAB bundle from gzipped tar file",
		Long:  importDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("this command requires the path to the packaged bundle")
			}
			importc.source = args[0]

			return importc.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&importc.dest, "destination", "d", "", "Location to unpack bundle")
	f.BoolVarP(&importc.verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func (im *importCmd) run() error {
	source, err := filepath.Abs(im.source)
	if err != nil {
		return err
	}

	dest, err := filepath.Abs(im.dest)
	if err != nil {
		return err
	}

	l := loader.NewLoader()
	imp, err := packager.NewImporter(source, dest, l, im.verbose)
	if err != nil {
		return err
	}
	return imp.Import()
}
