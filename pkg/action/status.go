package action

import (
	"fmt"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/driver"
)

// Status runs a status action on a CNAB bundle.
type Status struct {
	Driver driver.Driver
}

// Run executes a status action in an image
func (i *Status) Run(c *claim.Claim, creds credentials.Set) error {
	op := opFromClaim(claim.ActionStatus, c, creds)
	if !i.Driver.Handles(op.ImageType) {
		return fmt.Errorf("driver does not handle image type %s", op.ImageType)
	}
	return i.Driver.Run(op)
}
