package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

var globalUsage = `The CNAB installer
`

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "duffle",
		Short: globalUsage,
		Long:  globalUsage,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello World")
		},
	}

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
