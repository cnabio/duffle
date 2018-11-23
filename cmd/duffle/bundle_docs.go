package main

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/deis/duffle/pkg/driver"

	"github.com/russross/blackfriday"

	"github.com/samfoo/mdcat/renderer"

	"github.com/spf13/cobra"
)

type bundleDocsCmd struct {
	out        io.Writer
	bundleFile string
	insecure   bool
}

func newBundleDocsCmd(w io.Writer) *cobra.Command {
	bd := &bundleDocsCmd{out: w}

	cmd := &cobra.Command{
		Use:     "docs",
		Aliases: []string{"doc", "info"},
		Short:   "displays bundle documentation stored in the first Docker invocation image in /cnab/app/readme",
		RunE: func(cmd *cobra.Command, args []string) error {

			bundle, err := bundleFileOrArg1(args, bd.bundleFile)
			if err != nil {
				return err
			}

			bun, err := loadBundle(bundle, bd.insecure)
			if err != nil {
				return err
			}

			tmp, err := ioutil.TempFile("", "duffle-tmp-doc-")
			if err != nil {
				return fmt.Errorf("cannot create temp file for bundle doc: %v", err)
			}
			defer tmp.Close()

			var image string
			for _, ii := range bun.InvocationImages {
				if ii.ImageType != "docker" {
					continue
				}
				image = ii.Image
			}
			if image == "" {
				return fmt.Errorf("no docker type invocation image found")
			}

			err = driver.CopyFromContainer(image, "/cnab/README.md", tmp)
			if err != nil {
				return fmt.Errorf("cannot copy from container: %v", err)
			}

			doc, err := ioutil.ReadFile(tmp.Name())
			if err != nil {
				return fmt.Errorf("cannot read tmp file: %v", err)
			}

			extensions := 0 |
				blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
				blackfriday.EXTENSION_FENCED_CODE |
				blackfriday.EXTENSION_AUTOLINK |
				blackfriday.EXTENSION_STRIKETHROUGH |
				blackfriday.EXTENSION_SPACE_HEADERS |
				blackfriday.EXTENSION_HEADER_IDS |
				blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
				blackfriday.EXTENSION_DEFINITION_LISTS
			output := blackfriday.Markdown(doc, &renderer.Console{}, extensions)

			fmt.Printf("%s", output)

			return nil
		},
	}

	cmd.Flags().StringVarP(&bd.bundleFile, "file", "f", "", "path to bundle file display docs")
	cmd.Flags().BoolVarP(&bd.insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")

	return cmd
}
