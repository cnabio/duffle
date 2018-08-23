package main

import (
	"errors"
	"io"

	"github.com/spf13/cobra"
)

const usage = `This command will uninstall an installation of a CNAB bundle`

// TODO
func newUninstallCmd(w io.Writer) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: usage,
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("This command requires exactly 1 argument: the name of the installation to uninstall")
			}

			unimplemented("duffle uninstall")
			return nil
		},
	}

	return cmd
}
