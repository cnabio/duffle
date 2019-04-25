package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/pathmapping"
	"github.com/pivotal/image-relocation/pkg/registry"

	"github.com/deislabs/cnab-go/bundle"

	"github.com/deislabs/duffle/pkg/duffle/home"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	relocateDesc = `
Relocates any docker and oci images, including invocation images, referenced by a bundle, tags and pushes the images to
a registry, and creates a new bundle with an updated invocation images section and an updated image map.

The --repository-prefix flag determines the repositories for the relocated images.
Each image is tagged with a name starting with the given prefix and pushed to the repository.

For example, if the repository-prefix is example.com/user, the image istio/proxyv2 is relocated
to a name starting with example.com/user/ and pushed to a repository hosted by example.com.
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
	mapping        pathmapping.PathMapping
	registryClient registry.Client
}

func newRelocateCmd(w io.Writer) *cobra.Command {
	relocate := &relocateCmd{out: w}

	cmd := &cobra.Command{
		Use:   "relocate [INPUT-BUNDLE] [OUTPUT-BUNDLE]",
		Short: "relocate images in a CNAB bundle",
		Long:  relocateDesc,
		Example: `duffle relocate helloworld hellorelocated --repository-prefix example.com/user
duffle relocate path/to/bundle.json relocatedbundle --repository-prefix example.com/user --input-bundle-is-file
duffle relocate helloworld path/to/relocatedbundle.json --repository-prefix example.com/user --output-bundle-is-file`,
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
			relocate.registryClient = registry.NewRegistryClient()

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
	bun, err := r.setup()
	if err != nil {
		return err
	}

	if err := r.relocate(bun); err != nil {
		return err
	}

	return r.writeBundle(bun)
}

func (r *relocateCmd) relocate(bun *bundle.Bundle) error {
	// mutate the input bundle to become the output bundle
	if !r.outputBundleIsFile {
		bun.Name = r.outputBundle
	}

	for i := range bun.InvocationImages {
		ii := bun.InvocationImages[i]
		modified, err := r.relocateImage(&ii.BaseImage)
		if err != nil {
			return err
		}
		if modified {
			bun.InvocationImages[i] = ii
		}
	}

	for k := range bun.Images {
		im := bun.Images[k]
		modified, err := r.relocateImage(&im.BaseImage)
		if err != nil {
			return err
		}
		if modified {
			bun.Images[k] = im
		}
	}

	return nil
}

func (r *relocateCmd) relocateImage(i *bundle.BaseImage) (bool, error) {
	if !isOCI(i.ImageType) && !isDocker(i.ImageType) {
		return false, nil
	}
	// map the image name
	n, err := image.NewName(i.Image)
	if err != nil {
		return false, err
	}
	rn := r.mapping(r.repoPrefix, n)

	// tag/push the image to its new repository
	dig, err := r.registryClient.Copy(n, rn)
	if err != nil {
		return false, err
	}
	if i.Digest != "" && dig.String() != i.Digest {
		// should not happen
		return false, fmt.Errorf("digest of image %s not preserved: old digest %s; new digest %s", i.Image, i.Digest, dig.String())
	}

	// update the imagemap
	i.OriginalImage = i.Image
	i.Image = rn.String()
	return true, nil
}

func isOCI(imageType string) bool {
	return imageType == "" || imageType == "oci"
}

func isDocker(imageType string) bool {
	return imageType == "docker"
}

func (r *relocateCmd) setup() (*bundle.Bundle, error) {
	bundleFile, err := resolveBundleFilePath(r.inputBundle, r.home.String(), r.inputBundleIsFile)
	if err != nil {
		return nil, err
	}

	bun, err := loadBundle(bundleFile)
	if err != nil {
		return nil, err
	}

	if err = bun.Validate(); err != nil {
		return nil, err
	}

	return bun, nil
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
