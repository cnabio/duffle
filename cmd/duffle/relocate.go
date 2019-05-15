package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/pathmapping"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/imagestore"
	"github.com/deislabs/duffle/pkg/imagestore/builder"
	"github.com/deislabs/duffle/pkg/loader"
	"github.com/deislabs/duffle/pkg/packager"
	"github.com/deislabs/duffle/pkg/relocator"
)

const (
	relocateDesc = `
Relocates any docker and oci images, including invocation images, referenced by a bundle, tags and pushes the images to
a registry, and creates a new bundle with an updated invocation images section and an updated image map.

The --repository-prefix flag determines the repositories for the relocated images.
Each image is tagged with a name starting with the given prefix and pushed to the repository.

For example, if the repository-prefix is example.com/user, the image istio/proxyv2 is relocated
to a name starting with example.com/user/ and pushed to a repository hosted by example.com.

If a thick bundle is relocated, the images are loaded from the bundle instead of from their registries before being
tagged and pushed. The new bundle is a thin bundle, regardless of whether the input bundle was thick or thin.
`
	invalidRepositoryChars = ":@\" "
)

type relocateCmd struct {
	// args
	inputBundle  string
	outputBundle string

	// flags
	repoPrefix         string
	inputBundleIsFile  bool
	outputBundleIsFile bool

	// context
	home home.Home
	out  io.Writer

	// dependencies
	mapping           pathmapping.PathMapping
	imageStoreBuilder imagestore.Builder
	imageStore        imagestore.Store
}

func newRelocateCmd(w io.Writer) *cobra.Command {
	relocate := &relocateCmd{out: w}

	cmd := &cobra.Command{
		Use:   "relocate [INPUT-BUNDLE] [OUTPUT-BUNDLE]",
		Short: "relocate images in a CNAB bundle",
		Long:  relocateDesc,
		Example: `duffle relocate helloworld hellorelocated --repository-prefix example.com/user
duffle relocate path/to/bundle.json relocatedbundle --repository-prefix example.com/user --input-bundle-is-file
duffle relocate helloworld path/to/relocatedbundle.json --repository-prefix example.com/user --output-bundle-is-file
duffle relocate thick.tgz relocatedbundle.json --repository-prefix example.com/user --input-bundle-is-file --output-bundle-is-file`,
		Args: cobra.ExactArgs(2),
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
			relocate.outputBundle = args[1]

			relocate.home = home.Home(homePath())

			relocate.mapping = pathmapping.FlattenRepoPathPreserveTagDigest
			relocate.imageStoreBuilder = builder.NewLocatingBuilder()

			return relocate.run()
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&relocate.inputBundleIsFile, "input-bundle-is-file", "", false, "Indicates that the input bundle source is a file path")
	f.BoolVarP(&relocate.outputBundleIsFile, "output-bundle-is-file", "", false, "Indicates that the output bundle destination is a file path")
	f.StringVarP(&relocate.repoPrefix, "repository-prefix", "r", "", "Prefix for relocated image names")
	cmd.MarkFlagRequired("repository-prefix")

	return cmd
}

func (r *relocateCmd) run() error {
	rel, bun, cleanup, err := r.setup()
	if err != nil {
		return err
	}
	defer cleanup()

	if err := rel.Relocate(); err != nil {
		return err
	}

	return r.writeBundle(bun)
}

// The caller is responsible for running the returned cleanup function, which may delete the returned bundle.
func (r *relocateCmd) setup() (*relocator.Relocator, *bundle.Bundle, func(), error) {
	nop := func() {}
	dest := ""
	bundleFile, err := resolveBundleFilePath(r.inputBundle, r.home.String(), r.inputBundleIsFile)
	if err != nil {
		return nil, nil, nop, err
	}

	var bun *bundle.Bundle

	if strings.HasSuffix(bundleFile, ".tgz") {
		source, err := filepath.Abs(bundleFile)
		if err != nil {
			return nil, nil, nop, err
		}

		dest, err = ioutil.TempDir("", "duffle-relocate-unzip")
		if err != nil {
			return nil, nil, nop, err
		}

		l := loader.NewLoader()
		imp, err := packager.NewImporter(source, dest, l, false)
		if err != nil {
			return nil, nil, nop, err
		}
		dest, bun, err = imp.Unzip()
		if err != nil {
			return nil, nil, nop, err
		}
	} else {
		bun, err = loadBundle(bundleFile)
		if err != nil {
			return nil, nil, nop, err
		}
	}

	if err = bun.Validate(); err != nil {
		return nil, nil, nop, err
	}

	r.imageStore, err = r.imageStoreBuilder.ArchiveDir(dest).Build()
	if err != nil {
		return nil, nil, nop, err
	}

	// mutate the input bundle to become the output bundle
	if !r.outputBundleIsFile {
		bun.Name = r.outputBundle
	}

	mapping := func(i image.Name) image.Name {
		return pathmapping.FlattenRepoPathPreserveTagDigest(r.repoPrefix, i)
	}

	reloc, err := relocator.NewRelocator(bun, mapping, r.imageStore)
	if err != nil {
		return nil, nil, nop, err
	}

	return reloc, bun, func() { os.RemoveAll(dest) }, nil
}

func (r *relocateCmd) writeBundle(bf *bundle.Bundle) error {
	data, digest, err := marshalBundle(bf)
	if err != nil {
		return fmt.Errorf("cannot marshal bundle: %v", err)
	}

	if r.outputBundleIsFile {
		if err := ioutil.WriteFile(r.outputBundle, data, 0644); err != nil {
			return fmt.Errorf("cannot write bundle to %s: %v", r.outputBundle, err)
		}
		return nil
	}

	if err := ioutil.WriteFile(filepath.Join(r.home.Bundles(), digest), data, 0644); err != nil {
		return fmt.Errorf("cannot store bundle : %v", err)

	}

	// record the new bundle in repositories.json
	if err := recordBundleReference(r.home, bf.Name, bf.Version, digest); err != nil {
		return fmt.Errorf("cannot record bundle: %v", err)
	}

	return nil
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
