package action

import (
	"fmt"
	"strings"

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

func opFromClaim(action string, c *claim.Claim, creds credentials.Set) *driver.Operation {
	env, files := creds.Flatten()
	return &driver.Operation{
		Action:       action,
		Installation: c.Name,
		Parameters:   c.Parameters,
		Image:        c.Bundle.InvocationImage.Image,
		ImageType:    c.Bundle.InvocationImage.ImageType,
		Revision:     c.Revision,
		Environment:  conflateEnv(action, c, env),
		Files:        files,
	}
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
