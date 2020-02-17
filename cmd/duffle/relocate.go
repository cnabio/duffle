package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/loader"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/pathmapping"
	"github.com/pivotal/image-relocation/pkg/transport"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/cnabio/duffle/pkg/duffle/home"
	"github.com/cnabio/duffle/pkg/imagestore"
	"github.com/cnabio/duffle/pkg/imagestore/construction"
	"github.com/cnabio/duffle/pkg/packager"
	"github.com/cnabio/duffle/pkg/relocator"
)

const (
	relocateDesc = `
Relocates any docker and oci images, including invocation images, referenced by a bundle, tags and pushes the images to
a registry, and creates a relocation mapping JSON file.

The --repository-prefix flag determines the repositories for the relocated images.
Each image is tagged with a name starting with the given prefix and pushed to the repository.

For example, if the repository-prefix is example.com/user, the image istio/proxyv2 is relocated
to a name starting with example.com/user/ and pushed to a repository hosted by example.com.

The generated relocation mapping file maps the original image references to their relocated counterparts. This file is
an optional input to the install, upgrade, and run commands.
`
	invalidRepositoryChars = ":@\" "
)

type relocateCmd struct {
	// args
	inputBundle string

	// flags
	repoPrefix        string
	bundleIsFile      bool
	relocationMapping string
	skipTLSVerify     bool
	caCertPaths       []string

	// context
	home home.Home
	out  io.Writer

	// dependencies
	mapping               pathmapping.PathMapping
	transportConstructor  func([]string, bool) (*http.Transport, error)
	imageStoreConstructor imagestore.Constructor
	imageStore            imagestore.Store
}

func newRelocateCmd(w io.Writer) *cobra.Command {
	relocate := &relocateCmd{out: w}

	cmd := &cobra.Command{
		Use:   "relocate [INPUT-BUNDLE]",
		Short: "relocate images in a CNAB bundle",
		Long:  relocateDesc,
		Example: `duffle relocate helloworld --relocation-mapping path/to/relmap.json --repository-prefix example.com/user
duffle relocate path/to/bundle.json --relocation-mapping path/to/relmap.json --repository-prefix example.com/user --bundle-is-file`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// validate --repository-prefix if it is set, otherwise fall through so that cobra will report the missing flag in its usual manner
			if cmd.Flags().Changed("repository-prefix") {
				if err := validateRepository(relocate.repoPrefix); err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			relocate.inputBundle = args[0]

			relocate.home = home.Home(homePath())

			relocate.mapping = pathmapping.FlattenRepoPathPreserveTagDigest
			relocate.transportConstructor = transport.NewHttpTransport
			relocate.imageStoreConstructor = construction.NewLocatingConstructor()

			return relocate.run()
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&relocate.bundleIsFile, "bundle-is-file", "f", false, "Indicates that the input bundle source is a file path")
	f.StringVarP(&relocate.relocationMapping, "relocation-mapping", "m", "", "Path for output relocation mapping JSON file")
	cmd.MarkFlagRequired("relocation-mapping")
	f.StringVarP(&relocate.repoPrefix, "repository-prefix", "p", "", "Prefix for relocated image names")
	cmd.MarkFlagRequired("repository-prefix")
	f.StringSliceVarP(&relocate.caCertPaths, "ca-cert-path", "", nil, "Path to CA certificate for verifying registry TLS certificates (can be repeated for multiple certificates)")
	f.BoolVarP(&relocate.skipTLSVerify, "skip-tls-verify", "", false, "Skip TLS certificate verification for registries")

	return cmd
}

func (r *relocateCmd) run() error {
	relMap := make(map[string]string)

	rel, cleanup, err := r.setup()
	if err != nil {
		return err
	}
	defer cleanup()

	if err := rel.Relocate(relMap); err != nil {
		return err
	}

	return r.writeRelocationMapping(relMap)
}

func inferAndLoadBundle(bundleFile string) (*bundle.Bundle, string, error) {
	if strings.HasSuffix(bundleFile, ".tgz") {
		bun, dest, err := unzipBundle(bundleFile)
		if err != nil {
			return nil, "", err
		}
		return bun, dest, nil
	}
	bun, err := loadBundle(bundleFile)
	if err != nil {
		return nil, "", err
	}
	return bun, "", nil
}

// The caller is responsible for running the returned cleanup function, which may delete the returned bundle.
func (r *relocateCmd) setup() (*relocator.Relocator, func(), error) {
	nop := func() {}
	bundleFile, err := resolveBundleFilePath(r.inputBundle, r.home.String(), r.bundleIsFile)
	if err != nil {
		return nil, nop, err
	}

	bun, dest, err := inferAndLoadBundle(bundleFile)
	if err != nil {
		return nil, nop, err
	}

	if err = bun.Validate(); err != nil {
		return nil, nop, err
	}

	transport, err := r.transportConstructor(r.caCertPaths, r.skipTLSVerify)
	if err != nil {
		return nil, nop, err
	}

	if r.imageStore, err = r.imageStoreConstructor(
		imagestore.WithArchiveDir(dest),
		imagestore.WithTransport(transport),
	); err != nil {
		return nil, nop, err
	}

	mapping := func(i image.Name) image.Name {
		return pathmapping.FlattenRepoPathPreserveTagDigest(r.repoPrefix, i)
	}

	reloc, err := relocator.NewRelocator(bun, mapping, r.imageStore, r.out)
	if err != nil {
		return nil, nop, err
	}

	return reloc, func() { os.RemoveAll(dest) }, nil
}

func unzipBundle(bundleFile string) (*bundle.Bundle, string, error) {
	source, err := filepath.Abs(bundleFile)
	if err != nil {
		return nil, "", err
	}

	dest, err := ioutil.TempDir("", "duffle-relocate-unzip")
	if err != nil {
		return nil, "", err
	}

	l := loader.NewLoader()
	imp, err := packager.NewImporter(source, dest, l, false)
	if err != nil {
		return nil, "", err
	}
	dest, bun, err := imp.Unzip()
	if err != nil {
		return nil, "", err
	}

	return bun, dest, nil
}

func (r *relocateCmd) writeRelocationMapping(relMap map[string]string) error {
	rm, err := json.Marshal(relMap)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(r.relocationMapping, rm, 0644)
}

func validateRepository(repo string) error {
	if strings.HasSuffix(repo, "/") || strings.Contains(repo, "//") {
		return fmt.Errorf("invalid repository: trailing '/' and '//' not allowed: %s", repo)
	}

	for i, part := range strings.Split(repo, "/") {
		if i != 0 {
			if strings.ContainsAny(part, invalidRepositoryChars) {
				return fmt.Errorf("invalid repository: characters '%s' not allowed: %s", invalidRepositoryChars, repo)
			}
			continue
		}

		authorityParts := strings.Split(part, ":")
		if len(authorityParts) > 2 {
			return fmt.Errorf("invalid repository hostname: %s", part)
		}
		if errs := validation.IsDNS1123Subdomain(authorityParts[0]); len(errs) > 0 {
			return fmt.Errorf("invalid repository hostname: %s", strings.Join(errs, "; "))
		}
		if len(authorityParts) == 2 {
			portNumber, err := strconv.Atoi(authorityParts[1])
			if err != nil {
				return fmt.Errorf("invalid repository port number: %s", authorityParts[1])
			}

			if errs := validation.IsValidPortNum(portNumber); len(errs) > 0 {
				return fmt.Errorf("invalid repository port number: %s", strings.Join(errs, "; "))
			}
		}
	}

	return nil
}
