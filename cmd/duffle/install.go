package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/deis/duffle/pkg/action"
	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/loader"
	"github.com/deis/duffle/pkg/reference"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newInstallCmd() *cobra.Command {
	const usage = `Installs a CNAB bundle.

This installs a CNAB bundle with a specific installation name. Once the install is complete,
this bundle can be referenced by installation name.

Example:
	$ duffle install my_release duffle/example:0.1.0
	$ duffle status my_release

Different drivers are available for executing the duffle invocation image. The following drivers
are built-in:

	- docker: run the Docker client. Works for OCI and Docker images
	- debug: fake a run of the invocation image, and print out what would have been sent

Some drivers have additional configuration that can be passed via environment variable.

	docker:
	  - VERBOSE: "true" turns on extra output

UNIX Example:
	$ VERBOSE=true duffle install -d docker my_release duffle/example:0.1.0

Windows Example:
	$ $env:VERBOSE = true
	$ duffle install -d docker my_release duffle/example:0.1.0

For unpublished CNAB bundles, you can also load the bundle.json directly:

    $ duffle install dev_bundle -f path/to/bundle.json
`
	var (
		installDriver   string
		credentialsFile string
		valuesFile      string
		bundleFile      string
		setParams       []string
		insecure        bool
		setFiles        []string

		installationName string
		bun              *bundle.Bundle
	)

	cmd := &cobra.Command{
		Use:   "install NAME BUNDLE",
		Short: "install a CNAB bundle",
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			bundleFile, err := bundleFileOrArg2(args, bundleFile, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			installationName = args[0]

			bun, err = loadBundle(bundleFile, insecure)
			if err != nil {
				return err
			}

			if err = bun.Validate(); err != nil {
				return err
			}

			driverImpl, err := prepareDriver(installDriver)
			if err != nil {
				return err
			}

			creds, err := loadCredentials(credentialsFile, bun)
			if err != nil {
				return err
			}

			// Because this is an install, we create a new claim. For upgrades, we'd
			// load the claim based on installationName
			c, err := claim.New(installationName)
			if err != nil {
				return err
			}

			c.Bundle = bun
			c.Parameters, err = calculateParamValues(bun, valuesFile, setParams, setFiles)
			if err != nil {
				return err
			}

			inst := &action.Install{
				Driver: driverImpl,
			}
			fmt.Println("Executing install action...")
			err = inst.Run(c, creds, cmd.OutOrStdout())

			// Even if the action fails, we want to store a claim. This is because
			// we cannot know, based on a failure, whether or not any resources were
			// created. So we want to suggest that the user take investigative action.
			err2 := claimStorage().Store(*c)
			if err != nil {
				return fmt.Errorf("Install step failed: %v", err)
			}
			return err2
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")
	flags.StringVarP(&credentialsFile, "credentials", "c", "", "Specify a set of credentials to use inside the CNAB bundle")
	flags.StringVarP(&installDriver, "driver", "d", "docker", "Specify a driver name")
	flags.StringVarP(&valuesFile, "parameters", "p", "", "Specify file containing parameters. Formats: toml, MORE SOON")
	flags.StringVarP(&bundleFile, "file", "f", "", "Bundle file to install")
	flags.StringArrayVarP(&setParams, "set", "s", []string{}, "Set individual parameters as NAME=VALUE pairs")
	flags.StringArrayVarP(&setFiles, "set-file", "i", []string{}, "Set individual parameters from file content as NAME=SOURCE-PATH pairs")
	return cmd
}

func bundleFileOrArg2(args []string, bundleFile string, w io.Writer) (string, error) {
	switch {
	case len(args) < 1:
		return "", errors.New("This command requires at least one argument: NAME (name for the installation). It also requires a BUNDLE (CNAB bundle name) or file (using -f)\nValid inputs:\n\t$ duffle install NAME BUNDLE\n\t$ duffle install NAME -f path-to-bundle.json")
	case len(args) == 2 && bundleFile != "":
		return "", errors.New("please use either -f or specify a BUNDLE, but not both")
	case len(args) < 2 && bundleFile == "":
		return "", errors.New("required arguments are NAME (name of the installation) and BUNDLE (CNAB bundle name) or file")
	case len(args) == 2:
		return getBundleFile(args[1])
	}
	return bundleFile, nil
}

// optBundleFileOrArg2 optionally gets a bundle file.
// Returning an empty string with no error is a possible outcome.
func optBundleFileOrArg2(args []string, bundleFile string, w io.Writer) (string, error) {
	switch {
	case len(args) < 1:
		// No bundle provided
		return "", nil
	case len(args) == 2 && bundleFile != "":
		return "", errors.New("please use either -f or specify a BUNDLE, but not both")
	case len(args) < 2 && bundleFile == "":
		// No bundle provided
		return "", nil
	case len(args) == 2:
		var err error
		bundleFile, err = getBundleFile(args[1])
		if err != nil {
			return "", err
		}
	}
	return bundleFile, nil
}

// overrides parses the --set data and returns values that should override other params.
func overrides(overrides []string, paramDefs map[string]bundle.ParameterDefinition) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	for _, p := range overrides {
		pair := strings.SplitN(p, "=", 2)
		if len(pair) != 2 {
			// For now, I guess we skip cases where someone does --set foo or --set foo=
			// We could set this to an explicit nil and then use it as a trigger to unset
			// a parameter in the file.
			continue
		}
		def, ok := paramDefs[pair[0]]
		if !ok {
			return res, fmt.Errorf("parameter %s not defined in bundle", pair[0])
		}

		if _, ok := res[pair[0]]; ok {
			return res, fmt.Errorf("parameter %q specified multiple times", pair[0])
		}

		var err error
		res[pair[0]], err = def.ConvertValue(pair[1])
		if err != nil {
			return res, fmt.Errorf("can't use %s as value of %s: %s", pair[1], pair[0], err)
		}
	}
	return res, nil
}

func parseValues(file string) (map[string]interface{}, error) {
	v := viper.New()
	v.SetConfigFile(file)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return v.AllSettings(), nil
}

func getReference(bundleName string) (reference.NamedTagged, error) {
	var (
		name string
		ref  reference.NamedTagged
	)

	parts := strings.SplitN(bundleName, "://", 2)
	if len(parts) == 2 {
		name = parts[1]
	} else {
		name = parts[0]
	}
	normalizedRef, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image name: %s: %v", name, err)
	}
	if reference.IsNameOnly(normalizedRef) {
		ref, err = reference.WithTag(normalizedRef, "latest")
		if err != nil {
			// NOTE(bacongobbler): Using the default tag *must* be valid.
			// To create a NamedTagged type with non-validated
			// input, the WithTag function should be used instead.
			panic(err)
		}
	} else {
		if taggedRef, ok := normalizedRef.(reference.NamedTagged); ok {
			ref = taggedRef
		} else {
			return nil, fmt.Errorf("unsupported image name: %s", normalizedRef.String())
		}
	}

	return ref, nil
}

func getBundleRepoURL(bundleName string, home home.Home) (*url.URL, error) {
	ref, err := getReference(bundleName)
	if err != nil {
		return nil, err
	}

	proto := "https"
	parts := strings.Split(bundleName, "://")
	if len(parts) == 2 {
		proto = parts[0]
	}

	url := &url.URL{
		Scheme: proto,
		Host:   reference.Domain(ref),
		Path:   fmt.Sprintf("repositories/%s/tags/%s", reference.Path(ref), ref.Tag()),
	}
	return url, nil
}

func getBundleFile(bundleName string) (string, error) {
	home := home.Home(homePath())
	url, err := getBundleRepoURL(bundleName, home)
	if err != nil {
		return "", err
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request to %s responded with a non-200 status code: %d", url, resp.StatusCode)
	}

	bundle, err := bundle.ParseReader(resp.Body)
	if err != nil {
		return "", err
	}
	bundleFilepath := filepath.Join(home.Cache(), fmt.Sprintf("%s-%s.json", strings.Replace(bundle.Name, "/", "-", -1), bundle.Version))
	if err := bundle.WriteFile(bundleFilepath, 0644); err != nil {
		return "", err
	}

	return bundleFilepath, nil
}

func loadBundle(bundleFile string, insecure bool) (*bundle.Bundle, error) {
	var l loader.Loader
	if insecure {
		l = loader.NewUnsignedLoader()
	} else {
		kr, err := loadVerifyingKeyRings(homePath())
		if err != nil {
			return nil, err
		}
		l = loader.NewSecureLoader(kr)
	}
	return l.Load(bundleFile)
}

func calculateParamValues(bun *bundle.Bundle, valuesFile string, setParams, setFilePaths []string) (map[string]interface{}, error) {
	vals := map[string]interface{}{}
	if valuesFile != "" {
		var err error
		vals, err = parseValues(valuesFile)
		if err != nil {
			return vals, err
		}

	}
	overridden, err := overrides(setParams, bun.Parameters)
	if err != nil {
		return vals, err
	}
	for k, v := range overridden {
		vals[k] = v
	}

	// Now add files.
	for _, p := range setFilePaths {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return vals, fmt.Errorf("malformed set-file parameter: %q (must be NAME=PATH)", p)
		}

		// Check that this is a known param
		if _, ok := bun.Parameters[parts[0]]; !ok {
			return vals, fmt.Errorf("bundle does not have a parameter named %q", parts[0])
		}

		if _, ok := overridden[parts[0]]; ok {
			return vals, fmt.Errorf("parameter %q specified multiple times", parts[0])
		}
		content, err := ioutil.ReadFile(parts[1])
		if err != nil {
			return vals, fmt.Errorf("error while reading file %q: %s", parts[1], err)
		}
		vals[parts[0]] = string(content)
	}

	return bundle.ValuesOrDefaults(vals, bun)
}
