package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/action"
)

const usage = `This command will uninstall an installation of a CNAB bundle`

var uninstallDriver string

type uninstallCmd struct {
	out  io.Writer
	name string
}

func newUninstallCmd(w io.Writer) *cobra.Command {
	uc := &uninstallCmd{out: w}

	var credentialsFile string

	cmd := &cobra.Command{
		Use:          "uninstall",
		Short:        usage,
		Long:         usage,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("This command requires exactly 1 argument: the name of the installation to uninstall")
			}
			uc.name = args[0]

			return uc.uninstall(credentialsFile)
		},
	}

	cmd.Flags().StringVarP(&uninstallDriver, "driver", "d", "docker", "Specify a driver name")
	cmd.Flags().StringVarP(&credentialsFile, "credentials", "c", "", "Specify a set of credentials to use inside the CNAB bundle")

	return cmd
}

func (un *uninstallCmd) uninstall(credentialsFile string) error {

	claim, err := claimStorage().Read(un.name)
	if err != nil {
		return fmt.Errorf("%v not found: %v", un.name, err)
	}

	driverImpl, err := prepareDriver(uninstallDriver)
	if err != nil {
		return err
	}

	creds, err := loadCredentials(credentialsFile)
	if err != nil {
		return err
	}

	uninst := &action.Uninstall{
		Driver: driverImpl,
	}

	if err := uninst.Run(&claim, creds); err != nil {
		return fmt.Errorf("could not uninstall %q: %s", un.name, err)
	}
	return claimStorage().Delete(un.name)
}
