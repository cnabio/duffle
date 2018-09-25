package main

import (
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verbose bool

func newRootCmd(w io.Writer) *cobra.Command {
	const usage = `The CNAB installer`

	cmd := &cobra.Command{
		Use:          "duffle",
		Short:        usage,
		SilenceUsage: true,
		Long:         usage,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
		},
	}
	cmd.SetOutput(w)

	p := cmd.PersistentFlags()
	p.StringVar(&duffleHome, "home", defaultDuffleHome(), "location of your Duffle config. Overrides $DUFFLE_HOME")
	p.BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	cmd.AddCommand(newBuildCmd(w))
	cmd.AddCommand(newInitCmd(w))
	cmd.AddCommand(newListCmd(w))
	cmd.AddCommand(newPullCmd(w))
	cmd.AddCommand(newPushCmd(w))
	cmd.AddCommand(newRepoCmd(w))
	cmd.AddCommand(newSearchCmd(w))
	cmd.AddCommand(newVersionCmd(w))
	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newStatusCmd(w))
	cmd.AddCommand(newUninstallCmd())
	cmd.AddCommand(newUpgradeCmd())
	cmd.AddCommand(newCredentialsCmd(w))
	cmd.AddCommand(newRewriteCmd(w))

	return cmd
}
