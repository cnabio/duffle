package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/action"
)

const usage = `Uninstalls an installation of a CNAB bundle.

When using '--parameters' or '--set', the uninstall command will replace the old
parameters with the new ones supplied (even if the new set is an empty set). If neither
'--parameters' nor '--set' is passed, then the parameters used for 'duffle install' will
be re-used.
`

var uninstallDriver string

type uninstallCmd struct {
	duffleCmd
	name       string
	bundleFile string
	valuesFile string
	setParams  []string
}

func newUninstallCmd() *cobra.Command {
	uc := &uninstallCmd{}

	var (
		credentialsFile string
		bundleFile      string
	)

	cmd := &cobra.Command{
		Use:     "uninstall [NAME]",
		Short:   "uninstall CNAB installation",
		Long:    usage,
		PreRunE: uc.Prepare(),
		RunE: func(cmd *cobra.Command, args []string) error {
			uc.name = args[0]
			uc.Out = cmd.OutOrStdout()
			bundleFile, err := bundleFileOrArg2(args, bundleFile, uc.Out)
			// If no bundle was found, we just wait for the claim system
			// to load its bundleFile
			if err == nil {
				uc.bundleFile = bundleFile
			}

			return uc.uninstall(credentialsFile)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&uninstallDriver, "driver", "d", "docker", "Specify a driver name")
	flags.StringVarP(&credentialsFile, "credentials", "c", "", "Specify a set of credentials to use inside the CNAB bundle")
	flags.StringVarP(&bundleFile, "file", "f", "", "bundle file to install")
	flags.StringVarP(&uc.valuesFile, "parameters", "p", "", "Specify file containing parameters. Formats: toml, MORE SOON")
	flags.StringArrayVarP(&uc.setParams, "set", "s", []string{}, "set individual parameters as NAME=VALUE pairs")

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
		claim.Bundle = &b
	}

	// If no params are specified, allow re-use. But if params are set -- even if empty --
	// replace the existing params.
	if len(un.setParams) > 0 || un.valuesFile != "" {
		if claim.Bundle == nil {
			return errors.New("parameters can only be set if a bundle is provided")
		}
		params, err := calculateParamValues(claim.Bundle, un.valuesFile, un.setParams)
		if err != nil {
			return err
		}
		claim.Parameters = params
	}

	driverImpl, err := prepareDriver(uninstallDriver)
	if err != nil {
		return fmt.Errorf("could not prepare driver: %s", err)
	}

	creds, err := loadCredentials(credentialsFile, claim.Bundle)
	if err != nil {
		return fmt.Errorf("could not load credentials: %s", err)
	}

	uninst := &action.Uninstall{
		Driver: driverImpl,
	}

	fmt.Fprintln(un.Out, "Executing uninstall action...")
	if err := uninst.Run(&claim, creds, un.Out); err != nil {
		return fmt.Errorf("could not uninstall %q: %s", un.name, err)
	}
	return claimStorage().Delete(un.name)
}
