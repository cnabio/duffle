package main

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/packager"

	"github.com/spf13/cobra"
)

const exportDesc = `
Packages an invocation image together with all of its associated images and generates a single gzipped tarfile

All images specified in the bundle metadata are saved as tar files in the artifacts/ directory along with an artifacts.json file which describes the contents of artifacts/.
`

type exportCmd struct {
	dest string
	path string
	home home.Home
	out  io.Writer
	full bool
}

func newExportCmd(w io.Writer) *cobra.Command {
	export := &exportCmd{out: w}

	cmd := &cobra.Command{
		Use:   "export [PATH]",
		Short: "package CNAB bundle in gzipped tar file",
		Long:  exportDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("this command requires the path to the bundle")
			}
			export.home = home.Home(homePath())
			export.path = args[0]

			return export.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&export.dest, "destination", "d", "", "Save exported bundle to path")
	f.BoolVarP(&export.full, "full", "u", true, "Save bundle with all associated images")

	return cmd
}

func (ex *exportCmd) run() error {
	source, err := filepath.Abs(ex.path)
	if err != nil {
		return err
	}

	exp, err := packager.NewExporter(source, ex.dest, ex.full)
	if err != nil {
		return fmt.Errorf("Unable to set up exporter: %s", err)
	}
	if err := exp.Export(); err != nil {
		return err
	}
	return nil
}
