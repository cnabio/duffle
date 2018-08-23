package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/deis/duffle/pkg/action"
	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/driver"
	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/loader"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

//
func newInstallCmd(w io.Writer) *cobra.Command {
	const usage = `Install a CNAB bundle

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
	$ VERBOSE=true duffle install -d docker  install my_release duffle/example:0.1.0

Windows Example:
	$ $env:VERBOSE = true
	$ duffle install -d docker install my_release duffle/example:0.1.0
`
	var (
		installDriver string
		valuesFile    string
		bundleFile    string

		installationName string
		bundle           bundle.Bundle
	)

	cmd := &cobra.Command{
		Use:   "install NAME BUNDLE",
		Short: "install a CNAB bundle",
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) == 2 && bundleFile != "" {
				return errors.New("please use either -f or specify a BUNDLE, but not both")
			}

			if len(args) < 2 && bundleFile == "" {
				return errors.New("required arguments are NAME (name of the instllation) and BUNDLE (CNAB bundle name) or file")
			}

			if len(args) == 2 {
				// load bundleFile from a repository
				bundleName := args[1]
				relevantBundles := search([]string{bundleName})
				switch len(relevantBundles) {
				case 0:
					return fmt.Errorf("no bundles with the name '%s' was found", bundleName)
				case 1:
					bundleName = relevantBundles[0]
				default:
					var match bool
					// check if we have an exact match
					for _, f := range relevantBundles {
						if strings.Compare(f, bundleName) == 0 {
							bundleName = f
							match = true
						}
					}
					if !match {
						return fmt.Errorf("%d bundles with the name '%s' was found: %v", len(relevantBundles), bundleName, relevantBundles)
					}
				}
				filePath, repo, err := getBundleFile(bundleName)
				if err != nil {
					return err
				}
				bundleFile = filePath

				fmt.Fprintf(w, "loaded %s from repository %s\n", bundleFile, repo)
			}

			l, err := loader.New(bundleFile)
			if err != nil {
				return err
			}

			installationName = args[0]

			bundle, err = l.Load()
			if err != nil {
				return err
			}

			if err = validateImage(bundle.InvocationImage); err != nil {
				return err
			}

			driverImpl, err := prepareDriver(installDriver)
			if err != nil {
				return err
			}

			// Because this is an install, we create a new claim. For upgrades, we'd
			// load the claim based on installationName
			c := claim.New(installationName)
			c.Bundle = bundle.InvocationImage.Image
			c.ImageType = bundle.InvocationImage.ImageType
			if valuesFile != "" {
				vals, err := parseValues(valuesFile)
				if err != nil {
					return err
				}
				c.Parameters = vals
			}

			err = claimStorage().Store(*c)
			if err != nil {
				return err
			}

			inst := &action.Install{
				Driver: driverImpl,
			}
			return inst.Run(c)
		},
	}

	//cmd.Flags().StringSliceP("credentials", "c", []string{}, "Specify one or more credential sets")
	cmd.Flags().StringVarP(&installDriver, "driver", "d", "docker", "Specify a driver name")
	cmd.Flags().StringVarP(&valuesFile, "parameters", "p", "", "Specify file containing parameters. Formats: toml, MORE SOON")
	cmd.Flags().StringVarP(&bundleFile, "file", "f", "", "bundle file to install")
	return cmd
}

func prepareDriver(driverName string) (driver.Driver, error) {
	driverImpl, err := driver.Lookup(driverName)
	if err != nil {
		return driverImpl, err
	}

	// Load any driver-specific config out of the environment.
	if configurable, ok := driverImpl.(driver.Configurable); ok {
		driverCfg := map[string]string{}
		for env := range configurable.Config() {
			driverCfg[env] = os.Getenv(env)
		}
		configurable.SetConfig(driverCfg)
	}

	return driverImpl, err
}

func validateImage(img bundle.InvocationImage) error {
	switch img.ImageType {
	case "docker", "oci":
		return validateDockerish(img.Image)
	default:
		return nil
	}
}

func validateDockerish(s string) error {
	if !strings.Contains(s, ":") {
		return errors.New("version is required")
	}
	return nil
}

func parseValues(file string) (map[string]interface{}, error) {
	vals := map[string]interface{}{}
	ext := filepath.Ext(file)
	switch ext {
	case ".toml":
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return vals, err
		}
		err = toml.Unmarshal(data, &vals)
		return vals, err
	case ".json":
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return vals, err
		}
		err = json.Unmarshal(data, &vals)
		return vals, err
	default:
		return vals, errors.New("no decoder for " + ext)
	}
}

func getBundleFile(bundleName string) (string, string, error) {
	var (
		name string
		repo string
	)
	home := home.Home(homePath())
	bundleInfo := strings.Split(bundleName, "/")
	if len(bundleInfo) == 1 {
		name = bundleInfo[0]
		repo = home.DefaultRepository()
	} else {
		name = bundleInfo[len(bundleInfo)-1]
		repo = path.Dir(bundleName)
	}
	if strings.Contains(name, "./\\") {
		return "", "", fmt.Errorf("bundle name '%s' is invalid. Bundle names cannot include the following characters: './\\'", name)
	}

	return filepath.Join(home.Repositories(), repo, "bundles", fmt.Sprintf("%s.json", name)), repo, nil
}
