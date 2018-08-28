package main

import (
	"io"

	"github.com/spf13/cobra"
)

// TODO
func newRootCmd(w io.Writer) *cobra.Command {
	const usage = `The CNAB installer`

	cmd := &cobra.Command{
		Use:   "duffle",
		Short: usage,
		Long:  usage,
	}

	p := cmd.PersistentFlags()
	p.StringVar(&duffleHome, "home", defaultDuffleHome(), "location of your Duffle config. Overrides $DUFFLE_HOME")

	cmd.AddCommand(newBuildCmd(w))
	cmd.AddCommand(newInitCmd(w))
	cmd.AddCommand(newListCmd(w))
	cmd.AddCommand(newPullCmd(w))
	cmd.AddCommand(newPushCmd(w))
	cmd.AddCommand(newRepoCmd(w))
	cmd.AddCommand(newSearchCmd(w))
	cmd.AddCommand(newVersionCmd(w))
	cmd.AddCommand(newInstallCmd(w))
	cmd.AddCommand(newStatusCmd(w))
	cmd.AddCommand(newUninstallCmd(w))
	cmd.AddCommand(newUpgradeCmd(w))

	return cmd
}
