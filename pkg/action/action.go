package action

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/driver"
)

// Action describes one of the primary actions that can be executed in CNAB.
//
// The actions are:
// - install
// - upgrade
// - uninstall
// - downgrade
// - status
type Action interface {
	// Run an action, and record the status in the given claim
	Run(*claim.Claim, credentials.Set) error
}

func selectInvocationImage(d driver.Driver, c *claim.Claim) (bundle.InvocationImage, error) {
	if len(c.Bundle.InvocationImages) == 0 {
		return bundle.InvocationImage{}, errors.New("no invocationImages are defined in the bundle")
	}

	for _, ii := range c.Bundle.InvocationImages {
		if d.Handles(ii.ImageType) {
			return ii, nil
		}
	}

	return bundle.InvocationImage{}, errors.New("driver is not compatible with any of the invocation images in the bundle")
}

func opFromClaim(action string, c *claim.Claim, ii bundle.InvocationImage, creds credentials.Set, w io.Writer) (*driver.Operation, error) {
	env, files, err := creds.Expand(c.Bundle)
	for k, v := range c.Files {
		files[c.Bundle.Files[k].Path] = v
	}
	return &driver.Operation{
		Action:       action,
		Installation: c.Name,
		Parameters:   c.Parameters,
		Image:        ii.Image,
		ImageType:    ii.ImageType,
		Revision:     c.Revision,
		Environment:  conflateEnv(action, c, env),
		Files:        files,
		Out:          w,
	}, err
}

// conflateEnv combines all the stuff that should be placed into env vars
// It returns a new map.
func conflateEnv(action string, c *claim.Claim, env map[string]string) map[string]string {
	env["CNAB_INSTALLATION_NAME"] = c.Name
	env["CNAB_ACTION"] = action
	env["CNAB_BUNDLE_NAME"] = c.Bundle.Name
	env["CNAB_BUNDLE_VERSION"] = c.Bundle.Version

	for k, v := range c.Parameters {
		// TODO: Vet against bundle's parameters.json
		env[fmt.Sprintf("CNAB_P_%s", strings.ToUpper(k))] = fmt.Sprintf("%v", v)
	}
	return env
}
