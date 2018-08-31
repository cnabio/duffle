package action

import (
	"fmt"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/driver"
)

// Install describes an installation action
type Install struct {
	Driver driver.Driver // Needs to be more than a string
}

// Run performs an installation and updates the Claim accordingly
func (i *Install) Run(c *claim.Claim, creds credentials.Set) error {
	imageType := c.Bundle.InvocationImage.ImageType
	if !i.Driver.Handles(imageType) {
		return fmt.Errorf("driver does not handle image type %s", imageType)
	}

	op := opFromClaim(claim.ActionInstall, c, creds)
	if err := i.Driver.Run(op); err != nil {
		c.Update(claim.ActionInstall, claim.StatusFailure)
		c.Result.Message = err.Error()
		return err
	}

	// Update claim:
	c.Update(claim.ActionInstall, claim.StatusSuccess)
	return nil
}
