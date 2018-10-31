package action

import (
	"errors"
	"os"
	"strings"
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
	"secret_one": "I'm a secret",
	"secret_two": "I'm also a secret",
}

func (d *mockFailingDriver) Handles(imageType string) bool {
	return d.shouldHandle
}
func (d *mockFailingDriver) Run(op *driver.Operation) error {
	return errors.New("I always fail")
}

func mockBundle() *bundle.Bundle {
	return &bundle.Bundle{
		Name:    "bar",
		Version: "0.1.0",
		InvocationImages: []bundle.InvocationImage{
			{Image: "foo/bar:0.1.0", ImageType: "docker"},
		},
		Credentials: map[string]bundle.Location{
			"secret_one": {
				Path: "/foo/bar",
			},
			"secret_two": {
				EnvironmentVariable: "SECRET_TWO",
			},
		},
		Parameters: map[string]bundle.ParameterDefinition{
			"param_one": {
				DefaultValue: "one",
			},
			"param_two": {
				DefaultValue: "two",
				Destination: &bundle.Location{
					EnvironmentVariable: "PARAM_TWO",
				},
			},
			"param_three": {
				DefaultValue: "three",
				Destination: &bundle.Location{
					Path: "/param/three",
				},
			},
		},
	}

}

func TestOpFromClaim(t *testing.T) {
	now := time.Now()
	c := &claim.Claim{
		Created:  now,
		Modified: now,
		Name:     "name",
		Revision: "revision",
		Bundle:   mockBundle(),
		Parameters: map[string]interface{}{
			"param_one":   "oneval",
			"param_two":   "twoval",
			"param_three": "threeval",
		},
	}
	invocImage := c.Bundle.InvocationImages[0]

	op, err := opFromClaim(claim.ActionInstall, c, invocImage, mockSet, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	is := assert.New(t)

	is.Equal(c.Name, op.Installation)
	is.Equal(c.Revision, op.Revision)
	is.Equal(invocImage.Image, op.Image)
	is.Equal(driver.ImageTypeDocker, op.ImageType)
	is.Equal(op.Environment["SECRET_ONE"], "I'm a secret")
	is.Equal(op.Environment["PARAM_TWO"], "twoval")
	is.Equal(op.Environment["CNAB_P_PARAM_ONE"], "oneval")
	is.Equal(op.Files["secret_two"], "I'm also a secret")
	is.Equal(op.Files["/param/three"], "threeval")
	is.Len(op.Parameters, 3)
	is.Equal(os.Stdout, op.Out)
}

func TestSelectInvocationImage_EmptyInvocationImages(t *testing.T) {
	c := &claim.Claim{
		Bundle: &bundle.Bundle{},
	}
	_, err := selectInvocationImage(&driver.DebugDriver{}, c)
	if err == nil {
		t.Fatal("expected an error")
	}
	want := "no invocationImages are defined"
	got := err.Error()
	if !strings.Contains(got, want) {
		t.Fatalf("expected an error containing %q but got %q", want, got)
	}
}

func TestSelectInvocationImage_DriverIncompatible(t *testing.T) {
	c := &claim.Claim{
		Bundle: mockBundle(),
	}
	_, err := selectInvocationImage(&mockFailingDriver{}, c)
	if err == nil {
		t.Fatal("expected an error")
	}
	want := "driver is not compatible"
	got := err.Error()
	if !strings.Contains(got, want) {
		t.Fatalf("expected an error containing %q but got %q", want, got)
	}
}
