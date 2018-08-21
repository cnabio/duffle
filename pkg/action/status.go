package action

import (
	"fmt"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/driver"
)

// Status runs a status action on a CNAB bundle.
type Status struct {
	Driver driver.Driver
}

func (i *Status) Run(c *claim.Claim) error {
	// FIXME: Need to set op.ImageType and op.Credentials
	op := opFromClaim(c)
	if !i.Driver.Handles(op.ImageType) {
		return fmt.Errorf("driver does not handle image type %s", op.ImageType)
	}
	return i.Driver.Run(op)
}
