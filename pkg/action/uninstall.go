package action

import (
	"fmt"
	"io"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/driver"
)

// Uninstall runs an uninstall action
type Uninstall struct {
	Driver driver.Driver
}

// Run performs the uninstall steps and updates the Claim
func (u *Uninstall) Run(c *claim.Claim, creds credentials.Set, w io.Writer) error {
	op := opFromClaim(claim.ActionUninstall, c, creds, w)
	if !u.Driver.Handles(op.ImageType) {
		return fmt.Errorf("driver does not handle image type %s", op.ImageType)
	}
	if err := u.Driver.Run(op); err != nil {
		c.Update(claim.ActionUninstall, claim.StatusFailure)
		c.Result.Message = err.Error()
		return err
	}

	c.Update(claim.ActionUninstall, claim.StatusSuccess)
	return nil
}
