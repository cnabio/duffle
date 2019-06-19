package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	duffleDriver "github.com/deislabs/duffle/pkg/driver"
	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/loader"
	"github.com/deislabs/duffle/pkg/reference"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/credentials"
	"github.com/deislabs/cnab-go/driver"
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/spf13/cobra"
)

var (
	// duffleHome depicts the home directory where all duffle config is stored.
	duffleHome string
	rootCmd    *cobra.Command
)

func main() {
	rootCmd = newRootCmd(nil)
	must(rootCmd.Execute())
}

func homePath() string {
	return os.ExpandEnv(duffleHome)
}

func defaultDuffleHome() string {
	if home := os.Getenv(home.HomeEnvVar); home != "" {
		return home
	}

	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, ".duffle")
}

// claimStorage returns a claim store for accessing claims.
func claimStorage() claim.Store {
	h := home.Home(homePath())
	return claim.NewClaimStore(crud.NewFileSystemStore(h.Claims(), "json"))
}

// loadCredentials loads a set of credentials from HOME.
func loadCredentials(files []string, b *bundle.Bundle) (map[string]string, error) {
	creds := map[string]string{}
	if len(files) == 0 {
		return creds, credentials.Validate(creds, b.Credentials)
	}

	// The strategy here is "last one wins". We loop through each credential file and
	// calculate its credentials. Then we insert them into the creds map in the order
	// in which they were supplied on the CLI.
	for _, file := range files {
		if !isPathy(file) {
			file = filepath.Join(home.Home(homePath()).Credentials(), file+".yaml")
		}
		cset, err := credentials.Load(file)
		if err != nil {
			return creds, err
		}
		res, err := cset.Resolve()
		if err != nil {
			return res, err
		}

		for k, v := range res {
			creds[k] = v
		}
	}
	return creds, credentials.Validate(creds, b.Credentials)
}

// isPathy checks to see if a name looks like a path.
func isPathy(name string) bool {
	return strings.Contains(name, string(filepath.Separator))
}

func must(err error) {
	if err != nil {
		os.Exit(1)
	}
}

// prepareDriver prepares a driver per the user's request.
func prepareDriver(driverName string, relMap string) (driver.Driver, error) {
	driverImpl, err := duffleDriver.Lookup(driverName)
	if err != nil {
		return nil, err
	}

	// Load any driver-specific config out of the environment.
	if configurable, ok := driverImpl.(driver.Configurable); ok {
		driverCfg := map[string]string{}
		for env := range configurable.Config() {
			driverCfg[env] = os.Getenv(env)
		}
		configurable.SetConfig(driverCfg)
	}

	rm, err := loadRelMapping(relMap)
	if err != nil {
		return nil, err
	}

	// wrap the driver so any relocation mapping is mounted
	return &driverWithRelocationMapping{
		driver:     driverImpl,
		relMapping: rm,
	}, nil
}

type driverWithRelocationMapping struct {
	driver     driver.Driver
	relMapping string
}

func (d *driverWithRelocationMapping) Run(op *driver.Operation) error {
	// if there is a relocation mapping, ensure it is mounted
	if d.relMapping != "" {
		op.Files["/cnab/app/relocation-mapping.json"] = d.relMapping
	}
	return d.driver.Run(op)
}

func (d *driverWithRelocationMapping) Handles(it string) bool {
	return d.driver.Handles(it)
}

func loadRelMapping(relMap string) (string, error) {
	if relMap != "" {
		data, err := ioutil.ReadFile(relMap)
		if err != nil {
			return "", fmt.Errorf("failed to read relocation mapping from %s: %v", relMap, err)
		}
		return string(data), nil
	}

	return "", nil
}

func loadBundle(bundleFile string) (*bundle.Bundle, error) {
	l := loader.NewLoader()

	// Issue #439: Errors that come back from the loader can be
	// pretty opaque.
	var bun *bundle.Bundle
	bun, err := l.Load(bundleFile)
	if err != nil {
		return bun, fmt.Errorf("cannot load bundle: %s", err)
	}
	return bun, nil
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
		return nil, fmt.Errorf("%q is not a valid bundle name: %v", name, err)
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
