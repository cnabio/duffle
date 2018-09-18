package action

import (
	"errors"
	"testing"
	"time"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/credentials"
	"github.com/deis/duffle/pkg/driver"

	"github.com/stretchr/testify/assert"
)

type mockFailingDriver struct {
	shouldHandle bool
}

var mockSet = credentials.Set{
	"secret_one": {
		EnvVar: "SECRET_ONE",
		Value:  "I'm a secret",
	},
	"secret_two": {
		Path:  "secret_two",
		Value: "I'm also a secret",
	},
}

func (d *mockFailingDriver) Handles(imageType string) bool {
	return d.shouldHandle
}
func (d *mockFailingDriver) Run(op *driver.Operation) error {
	return errors.New("I always fail")
}

func mockBundle() *bundle.Bundle {
	return &bundle.Bundle{
		Name:            "bar",
		Version:         "0.1.0",
		InvocationImage: bundle.InvocationImage{Image: "foo/bar:0.1.0", ImageType: "docker"},
	}

}

func TestOpFromClaim(t *testing.T) {
	now := time.Now()
	c := &claim.Claim{
		Created:    now,
		Modified:   now,
		Name:       "name",
		Revision:   "revision",
		Bundle:     mockBundle(),
		Parameters: map[string]interface{}{"duff": "beer"},
	}

	op := opFromClaim(claim.ActionInstall, c, mockSet)

	is := assert.New(t)

	is.Equal(c.Name, op.Installation)
	is.Equal(c.Revision, op.Revision)
	is.Equal(c.Bundle.InvocationImage.Image, op.Image)
	is.Equal(driver.ImageTypeDocker, op.ImageType)
	is.Equal(op.Environment["SECRET_ONE"], "I'm a secret")
	is.Equal(op.Files["secret_two"], "I'm also a secret")
	is.Len(op.Parameters, 1)
}
