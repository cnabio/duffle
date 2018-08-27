package main

import (
	"errors"
	"io"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/action"
	"github.com/deis/duffle/pkg/claim"
)

func newStatusCmd(w io.Writer) *cobra.Command {
	const short = "get the status of an installation"
	const long = `Get the status of an existing installation.

Given an installation name, execute the status task for this. A status
action will restart the CNAB image and ask it to query for status. For that
reason, it may need the same credentials used to install.
`
	var (
		statusDriver    string
		credentialsFile string
	)

	cmd := &cobra.Command{
		Use:   "status NAME",
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("required arg is NAME (installation name")
			}
			claimName := args[0]
			c, err := loadClaim(claimName)
			if err != nil {
				return err
			}

			creds, err := loadCredentials(credentialsFile)
			if err != nil {
				return err
			}

			driverImpl, err := prepareDriver(statusDriver)
			if err != nil {
				return err
			}

			// TODO: Do we pass new values in here? Or just from Claim?
			action := &action.Status{Driver: driverImpl}
			return action.Run(&c, creds)
		},
	}
	cmd.Flags().StringVarP(&statusDriver, "driver", "d", "docker", "Specify a driver name")
	cmd.Flags().StringVarP(&credentialsFile, "credentials", "c", "", "Specify a set of credentials to use inside the CNAB bundle")

	return cmd
}

func loadClaim(name string) (claim.Claim, error) {
	storage := claimStorage()
	return storage.Read(name)
}
