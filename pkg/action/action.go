package action

import (
	"github.com/deis/duffle/pkg/claim"
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
	Run(*claim.Claim) error
}

func opFromClaim(c *claim.Claim) *driver.Operation {
	return &driver.Operation{
		Action:       claim.ActionStatus,
		Installation: c.Name,
		Parameters:   c.Parameters,
		Credentials:  []driver.ResolvedCred{},
		Image:        c.Bundle,
		ImageType:    driver.ImageTypeDocker,
		Revision:     c.Revision,
	}
}
