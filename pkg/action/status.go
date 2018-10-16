package action

import (
	"io"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/driver"
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

	op := opFromClaim(claim.ActionStatus, c, invocImage, creds, w)
	return i.Driver.Run(op)
}
