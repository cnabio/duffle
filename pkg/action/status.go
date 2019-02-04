package action

import (
	"io"

	"github.com/scothis/ruffle/pkg/claim"
	"github.com/scothis/ruffle/pkg/credentials"
	"github.com/scothis/ruffle/pkg/driver"
)

// Status runs a status action on a CNAB bundle.
type Status struct {
	Driver driver.Driver
}

// Run executes a status action in an image
func (i *Status) Run(c *claim.Claim, creds credentials.Set, w io.Writer) error {
	invocImage, err := selectInvocationImage(i.Driver, c)
	if err != nil {
		return err
	}

	op, err := opFromClaim(claim.ActionStatus, c, invocImage, creds, w)
	if err != nil {
		return err
	}
	return i.Driver.Run(op)
}
