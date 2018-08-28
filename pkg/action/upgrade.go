package action

import (
	"fmt"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/driver"
)

// Upgrade runs an upgrade action
type Upgrade struct {
	Driver driver.Driver
}

// Run performs the upgrade steps and updates the Claim
func (u *Upgrade) Run(c *claim.Claim, creds credentials.Set) error {
	op := opFromClaim(claim.ActionUpgrade, c, creds)
	if !u.Driver.Handles(op.ImageType) {
		return fmt.Errorf("driver does not handle image type %s", op.ImageType)
	}
	if err := u.Driver.Run(op); err != nil {
		c.Update(claim.ActionUpgrade, claim.StatusFailure)
		c.Result.Message = err.Error()
		return err
	}

	c.Update(claim.ActionUpgrade, claim.StatusSuccess)
	return nil
}
