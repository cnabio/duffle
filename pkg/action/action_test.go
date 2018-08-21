package action

import (
	"errors"
	"testing"
	"time"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/driver"
	"github.com/stretchr/testify/assert"
)

type mockFailingDriver struct {
	shouldHandle bool
}

func (d *mockFailingDriver) Handles(imageType string) bool {
	return d.shouldHandle
}
func (d *mockFailingDriver) Run(op *driver.Operation) error {
	return errors.New("I always fail")
}

func TestOpFromClaim(t *testing.T) {
	now := time.Now()
	c := &claim.Claim{
		Created:    now,
		Modified:   now,
		Name:       "name",
		Revision:   "revision",
		Bundle:     "foo/bar:0.1.0",
		Parameters: map[string]interface{}{"duff": "beer"},
	}

	op := opFromClaim(c)

	is := assert.New(t)

	is.Equal(c.Name, op.Installation)
	is.Equal(c.Revision, op.Revision)
	is.Equal(c.Bundle, op.Image)
	is.Equal(driver.ImageTypeDocker, op.ImageType)
	is.Empty(op.Credentials)
	is.Len(op.Parameters, 1)

}
