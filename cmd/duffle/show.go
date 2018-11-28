package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

func newShowCmd(w io.Writer) *cobra.Command {
	var (
		insecure bool
		raw      bool
	)

	const usage = ` Returns information about an application bundle.

	Example:
		$ duffle show duffle/example:0.1.0

	To display unsigned bundles, pass the --insecure flag:
		$ duffle show duffle/unsinged-example:0.1.0 --insecure
`

	cmd := &cobra.Command{
		Use:   "show NAME",
		Short: "return low-level information on application bundles",
		Long:  usage,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundleName := args[0]

			bundleFile, err := getBundleFilepath(bundleName, insecure)
			if err != nil {
				return err
			}

			if raw {
				f, err := os.Open(bundleFile)
				if err != nil {
					return err
				}
				defer f.Close()
				_, err = io.Copy(w, f)
				return err
			}

			bun, err := loadBundle(bundleFile, insecure)
			if err != nil {
				return err
			}

			_, err = bun.WriteTo(w)

			return err
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")
	flags.BoolVarP(&raw, "raw", "r", false, "Display the raw bundle manifest")

	return cmd
}
