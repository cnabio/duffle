package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/repo"
)

func newInspectCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "return low-level information on application bundles",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := home.Home(homePath())
			bundleName := args[0]

			ref, err := getReference(bundleName)
			if err != nil {
				return fmt.Errorf("could not parse reference for %s: %v", bundleName, err)
			}

			// read the bundle reference from repositories.json
			index, err := repo.LoadIndex(home.Repositories())
			if err != nil {
				return fmt.Errorf("cannot open %s: %v", home.Repositories(), err)
			}

			digest, err := index.Get(ref.Name(), ref.Tag())
			if err != nil {
				return err
			}

			f, err := os.Open(filepath.Join(home.Bundles(), digest))
			if err != nil {
				return err
			}
			defer f.Close()

			io.Copy(w, f)

			return nil
		},
	}

	return cmd
}
