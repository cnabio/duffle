package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/deis/duffle/pkg/action"
	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/driver"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

//
func newInstallCmd(w io.Writer) *cobra.Command {
	const usage = `Install a CNAB package

This installs a CNAB package with a specific installation name. Once the install is complete,
this package can be referenced by installation name.

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
	var installDriver string
	var valuesFile string

	cmd := &cobra.Command{
		Use:   "install NAME BUNDLE",
		Short: "install a CNAB package",
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Basically, we want an image and a name. Then we want to create an installation
			// from that image, and record it with the name.
			if len(args) != 2 {
				return errors.New("required arguments are NAME (name of the instllation) and BUNDLE (CNAB bundle name)")
			}
			installationName := args[0]
			bundle := args[1]
			if err := validateDockerish(bundle); err != nil {
				return err
			}

			driverImpl, err := prepareDriver(installDriver)
			if err != nil {
				return err
			}

			// Because this is an install, we create a new claim. For upgrades, we'd
			// load the claim based on installationName
			c := claim.New(installationName)
			c.Bundle = bundle
			if valuesFile != "" {
				vals, err := parseValues(valuesFile)
				if err != nil {
					return err
				}
				c.Parameters = vals
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

	return cmd
}

func prepareDriver(driverName string) (driver.Driver, error) {
	driverImpl, err := driver.Lookup(driverName)
	if err != nil {
		return driverImpl, err
	}

	// Load any driver-specific config out of the environment.
	driverCfg := map[string]string{}
	for env := range driverImpl.Config() {
		driverCfg[env] = os.Getenv(env)
	}
	driverImpl.SetConfig(driverCfg)
	return driverImpl, err
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
