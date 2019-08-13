package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/deislabs/cnab-go/claim"

	"github.com/spf13/cobra"
)

const claimsShowDesc = `
Display the content of a claim.

This dumps the entire content of a claim as a JSON object.
`

type claimsShowCmd struct {
	Name       string
	OnlyBundle bool
	Output     string
	Storage    claim.Store
}

func newClaimsShowCmd(w io.Writer) *cobra.Command {
	var cmdData claimsShowCmd
	cmd := &cobra.Command{
		Use:     "show NAME",
		Short:   "show a claim",
		Long:    claimsShowDesc,
		Aliases: []string{"get"},
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cmdData.validateShowFlags()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdData.Name = args[0]
			cmdData.Storage = claimStorage()
			return cmdData.runClaimShow(w)
		},
	}

	cmd.Flags().BoolVarP(&cmdData.OnlyBundle, "bundle", "b", false, "only show the bundle from the claim")
	cmd.Flags().StringVarP(&cmdData.Output, "output", "o", "", "show the contents of the named output")

	return cmd
}

func (csc claimsShowCmd) validateShowFlags() error {
	if csc.OnlyBundle && csc.Output != "" {
		return errors.New("invalid flags: at most one of --bundle and --output can be specified")
	}
	return nil
}

func (csc claimsShowCmd) runClaimShow(w io.Writer) error {
	c, err := csc.Storage.Read(csc.Name)
	if err != nil {
		return err
	}

	if csc.OnlyBundle {
		return displayAsJSON(w, c.Bundle)
	}

	if csc.Output != "" {
		output, found := c.Outputs[csc.Output]
		if !found {
			return fmt.Errorf("unknown output name: %s", csc.Output)
		}

		_, err := fmt.Fprint(w, output)
		return err
	}

	return displayAsJSON(w, c)
}

func displayAsJSON(out io.Writer, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	out.Write([]byte("\n"))
	return err
}
