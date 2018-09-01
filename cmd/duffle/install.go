package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/deis/duffle/pkg/action"
	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/loader"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

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

		installationName string
		bundle           bundle.Bundle
	)

	cmd := &cobra.Command{
		Use:   "install NAME BUNDLE",
		Short: "install a CNAB bundle",
		Long:  usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case len(args) < 1:
				return errors.New("This command requires at least one argument: NAME (name for the installation). It also requires a BUNDLE (CNAB bundle name) or file (using -f)\nValid inputs:\n\t$ duffle install NAME BUNDLE\n\t$ duffle install NAME -f path-to-bundle.json")
			case len(args) == 2 && bundleFile != "":
				return errors.New("please use either -f or specify a BUNDLE, but not both")
			case len(args) < 2 && bundleFile == "":
				return errors.New("required arguments are NAME (name of the installation) and BUNDLE (CNAB bundle name) or file")
			case len(args) == 2:
				var err error
				bundleFile, err = findBundleJS(args[1], w)
				if err != nil {
					return err
				}
			}

			l, err := loader.New(bundleFile)
			if err != nil {
				return err
			}

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

			creds, err := loadCredentials(credentialsFile)
			if err != nil {
				return err
			}

			// Because this is an install, we create a new claim. For upgrades, we'd
			// load the claim based on installationName
			c, err := claim.New(installationName)
			if err != nil {
				return err
			}
			c.Bundle = bundle.InvocationImage.Image
			c.ImageType = bundle.InvocationImage.ImageType
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
			fmt.Println("Executing install action...")
			err = inst.Run(c, creds)

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

	cmd.Flags().StringVarP(&credentialsFile, "credentials", "c", "", "Specify a set of credentials to use inside the CNAB bundle")
	cmd.Flags().StringVarP(&installDriver, "driver", "d", "docker", "Specify a driver name")
	cmd.Flags().StringVarP(&valuesFile, "parameters", "p", "", "Specify file containing parameters. Formats: toml, MORE SOON")
	cmd.Flags().StringVarP(&bundleFile, "file", "f", "", "bundle file to install")
	return cmd
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

// findBundleJS tries to find the JS file by search the repo index
func findBundleJS(bundleName string, w io.Writer) (string, error) {
	relevantBundles := search([]string{bundleName})
	switch len(relevantBundles) {
	case 0:
		return bundleName, fmt.Errorf("no bundles with the name '%s' was found", bundleName)
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
			return bundleName, fmt.Errorf("%d bundles with the name '%s' were found: %v", len(relevantBundles), bundleName, relevantBundles)
		}
	}
	filePath, repo, err := getBundleFile(bundleName)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(w, "loaded %s from repository %s\n", filePath, repo)
	return filePath, nil
}
