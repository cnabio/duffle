package action

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/claim"
	"github.com/deislabs/duffle/pkg/credentials"
	"github.com/deislabs/duffle/pkg/driver"

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
			{
				bundle.BaseImage{Image: "foo/bar:0.1.0", ImageType: "docker"},
			},
		},
		Credentials: map[string]bundle.Location{
			"secret_one": {
				EnvironmentVariable: "SECRET_ONE",
				Path:                "/foo/bar",
			},
			"secret_two": {
				EnvironmentVariable: "SECRET_TWO",
				Path:                "/secret/two",
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
		Actions: map[string]bundle.Action{
			"test": {Modifies: true},
		},
		Images: map[string]bundle.Image{
			"image-a": {
				BaseImage: bundle.BaseImage{
					Image: "foo/bar:0.1.0", ImageType: "docker",
				},
				Description: "description",
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

	op, err := opFromClaim(claim.ActionInstall, stateful, c, invocImage, mockSet, os.Stdout)
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
	is.Equal(op.Files["/secret/two"], "I'm also a secret")
	is.Equal(op.Files["/param/three"], "threeval")
	is.Contains(op.Files, "/cnab/app/image-map.json")
	var imgMap map[string]bundle.Image
	is.NoError(json.Unmarshal([]byte(op.Files["/cnab/app/image-map.json"]), &imgMap))
	is.Equal(c.Bundle.Images, imgMap)
	is.Len(op.Parameters, 3)
	is.Equal(os.Stdout, op.Out)
}

func TestOpFromClaim_UndefinedParams(t *testing.T) {
	now := time.Now()
	c := &claim.Claim{
		Created:  now,
		Modified: now,
		Name:     "name",
		Revision: "revision",
		Bundle:   mockBundle(),
		Parameters: map[string]interface{}{
			"param_one":         "oneval",
			"param_two":         "twoval",
			"param_three":       "threeval",
			"param_one_million": "this is not a valid parameter",
		},
	}
	invocImage := c.Bundle.InvocationImages[0]

	_, err := opFromClaim(claim.ActionInstall, stateful, c, invocImage, mockSet, os.Stdout)
	assert.Error(t, err)
}

func TestOpFromClaim_MissingRequiredParameter(t *testing.T) {
	now := time.Now()
	b := mockBundle()
	b.Parameters["param_one"] = bundle.ParameterDefinition{Required: true}

	c := &claim.Claim{
		Created:  now,
		Modified: now,
		Name:     "name",
		Revision: "revision",
		Bundle:   b,
		Parameters: map[string]interface{}{
			"param_two":   "twoval",
			"param_three": "threeval",
		},
	}
	invocImage := c.Bundle.InvocationImages[0]

	// missing required parameter fails
	_, err := opFromClaim(claim.ActionInstall, stateful, c, invocImage, mockSet, os.Stdout)
	assert.EqualError(t, err, `missing required parameter "param_one" for action "install"`)

	// fill the missing parameter
	c.Parameters["param_one"] = "oneval"
	_, err = opFromClaim(claim.ActionInstall, stateful, c, invocImage, mockSet, os.Stdout)
	assert.Nil(t, err)
}

func TestOpFromClaim_MissingRequiredParamSpecificToAction(t *testing.T) {
	now := time.Now()
	b := mockBundle()
	// Add a required parameter only defined for the test action
	b.Parameters["param_test"] = bundle.ParameterDefinition{
		ApplyTo:  []string{"test"},
		Required: true,
	}
	c := &claim.Claim{
		Created:  now,
		Modified: now,
		Name:     "name",
		Revision: "revision",
		Bundle:   b,
		Parameters: map[string]interface{}{
			"param_one":   "oneval",
			"param_two":   "twoval",
			"param_three": "threeval",
		},
	}
	invocImage := c.Bundle.InvocationImages[0]

	// calling install action without the test required parameter for test action is ok
	_, err := opFromClaim(claim.ActionInstall, stateful, c, invocImage, mockSet, os.Stdout)
	assert.Nil(t, err)

	// test action needs the required parameter
	_, err = opFromClaim("test", stateful, c, invocImage, mockSet, os.Stdout)
	assert.EqualError(t, err, `missing required parameter "param_test" for action "test"`)

	c.Parameters["param_test"] = "only for test action"
	_, err = opFromClaim("test", stateful, c, invocImage, mockSet, os.Stdout)
	assert.Nil(t, err)
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
