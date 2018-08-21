package action

import (
	"fmt"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/driver"
)

// Install describes an installation action
type Install struct {
	Driver driver.Driver // Needs to be more than a string
}

// Run performs an installation and updates the Claim accordingly
func (i *Install) Run(c *claim.Claim) error {

	// TODO: We need to get the manifest so that we can check the image type. The
	// current theory is that we can store the manifest in the Docker registry and fetch
	// it without fetching the image. But we could also pass it in the CLI for non-Docker
	// images (like VMs)

	// TODO: load CNAB and resolve params
	// TODO: get image type from bundle.json and call driver.Handles() on that.
	// TODO: should credentials hang off Claim or be injected into Run() above

	op := opFromClaim(c)

	if !i.Driver.Handles(op.ImageType) {
		return fmt.Errorf("driver does not handle image type %s", op.ImageType)
	}

	// Perform install:
	if err := i.Driver.Run(op); err != nil {
		c.Update(claim.ActionInstall, claim.StatusFailure)
		c.Result.Message = err.Error()
		return err
	}

	// Update claim:
	c.Update(claim.ActionInstall, claim.StatusSuccess)
	return nil
}
