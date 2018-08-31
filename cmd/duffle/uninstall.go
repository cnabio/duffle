package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/action"
)

const usage = `This command will uninstall an installation of a CNAB bundle`

var uninstallDriver string

type uninstallCmd struct {
	out        io.Writer
	name       string
	bundleFile string
}

func newUninstallCmd(w io.Writer) *cobra.Command {
	uc := &uninstallCmd{out: w}

	var (
		credentialsFile string
		bundleFile      string
	)

	cmd := &cobra.Command{
		Use:   "uninstall [NAME]",
		Short: "uninstall CNAB installation",
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			uc.name = args[0]
			bundleFile, err := bundleFileOrArg2(args, bundleFile, w)
			// If no bundle was found, we just wait for the claim system
			// to load its bundleFile
			if err == nil {
				uc.bundleFile = bundleFile
			}

			return uc.uninstall(credentialsFile)
		},
	}

	cmd.Flags().StringVarP(&uninstallDriver, "driver", "d", "docker", "Specify a driver name")
	cmd.Flags().StringVarP(&credentialsFile, "credentials", "c", "", "Specify a set of credentials to use inside the CNAB bundle")
	cmd.Flags().StringVarP(&bundleFile, "file", "f", "", "bundle file to install")

	return cmd
}

func (un *uninstallCmd) uninstall(credentialsFile string) error {

	claim, err := claimStorage().Read(un.name)
	if err != nil {
		return fmt.Errorf("%v not found: %v", un.name, err)
	}

	if un.bundleFile != "" {
		b, err := loadBundle(un.bundleFile)
		if err != nil {
			return err
		}
		claim.Bundle = b.InvocationImage.Image
		claim.ImageType = b.InvocationImage.ImageType
	}

	driverImpl, err := prepareDriver(uninstallDriver)
	if err != nil {
		return err
	}

	creds, err := loadCredentials(credentialsFile, claim.Bundle)
	if err != nil {
		return err
	}

	uninst := &action.Uninstall{
		Driver: driverImpl,
	}

	fmt.Fprintf(un.out, "Executing uninstall action...")
	if err := uninst.Run(&claim, creds); err != nil {
		return fmt.Errorf("could not uninstall %q: %s", un.name, err)
	}
	return claimStorage().Delete(un.name)
}
