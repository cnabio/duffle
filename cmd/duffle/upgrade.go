package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/action"
)

const upgradeUsage = `This command will perform the upgrade action in the CNAB bundle`

var upgradeDriver string

type upgradeCmd struct {
	out  io.Writer
	name string
}

func newUpgradeCmd(w io.Writer) *cobra.Command {
	uc := &upgradeCmd{out: w}

	var credentialsFile string

	cmd := &cobra.Command{
		Use:   "upgrade [NAME]",
		Short: "upgrade CNAB installation",
		Long:  upgradeUsage,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("This command requires exactly 1 argument: the name of the installation to upgrade")
			}
			uc.name = args[0]

			return uc.upgrade(credentialsFile)
		},
	}

	cmd.Flags().StringVarP(&upgradeDriver, "driver", "d", "docker", "Specify a driver name")
	cmd.Flags().StringVarP(&credentialsFile, "credentials", "c", "", "Specify a set of credentials to use inside the CNAB bundle")

	return cmd
}

func (up *upgradeCmd) upgrade(credentialsFile string) error {

	claim, err := claimStorage().Read(up.name)
	if err != nil {
		return fmt.Errorf("%v not found: %v", up.name, err)
	}

	driverImpl, err := prepareDriver(upgradeDriver)
	if err != nil {
		return err
	}

	creds, err := loadCredentials(credentialsFile)
	if err != nil {
		return err
	}

	upgr := &action.Upgrade{
		Driver: driverImpl,
	}

	if err := upgr.Run(&claim, creds); err != nil {
		return fmt.Errorf("could not upgrade %q: %s", up.name, err)
	}

	return nil
}
