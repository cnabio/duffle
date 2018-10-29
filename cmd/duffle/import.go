package main

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/deis/duffle/pkg/packager"

	"github.com/spf13/cobra"
)

const importDesc = `
Unpacks a bundle from a gzipped tar file on local file system
`

type importCmd struct {
	source string
	dest   string
	out    io.Writer
}

func newImportCmd(w io.Writer) *cobra.Command {
	importc := &importCmd{out: w}

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
	f.StringVarP(&importc.dest, "destination", "d", "", "location to unpack bundle")

	return cmd
}

func (im *importCmd) run() error {
	source, err := filepath.Abs(im.source)
	if err != nil {
		return err
	}

	dest, err := filepath.Abs(im.dest) //TODO: double check
	if err != nil {
		return err
	}

	imp, err := packager.NewImporter(source, dest)
	if err != nil {
		return err
	}
	return imp.Import()
}
