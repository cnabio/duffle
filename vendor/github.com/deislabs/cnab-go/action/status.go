package action

import (
	"io"

	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/credentials"
	"github.com/deislabs/cnab-go/driver"
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

	op, err := opFromClaim(claim.ActionStatus, stateful, c, invocImage, creds, w)
	if err != nil {
		return err
	}

	// Ignore OperationResult because non-modifying actions don't have outputs to save.
	_, err = i.Driver.Run(op)
	return err
}
