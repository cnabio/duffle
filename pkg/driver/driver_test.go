package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ Driver = &DockerDriver{}
var _ Driver = &DebugDriver{}

func TestLookup_Error(t *testing.T) {
	d, err := Lookup("no_such_driver")

	assert.NotNil(t, d)
	assert.Error(t, err)
}

func TestDebugDriver_Handles(t *testing.T) {
	d, err := Lookup("debug")
	is := assert.New(t)
	is.NoError(err)
	is.NotNil(d)
	is.True(d.Handles(ImageTypeDocker))
	is.True(d.Handles("anything"))
}

func TestDebugDriver_Run(t *testing.T) {
	d, err := Lookup("debug")
	is := assert.New(t)
	is.NoError(err)
	is.NotNil(d)

	op := &Operation{
		Installation: "test",
		Image:        "test:1.2.3",
		ImageType:    "oci",
	}
	is.NoError(d.Run(op))
}

func TestDockerDriver_Handles(t *testing.T) {
	d, err := Lookup("docker")
	is := assert.New(t)
	is.NoError(err)
	is.NotNil(d)
	is.True(d.Handles(ImageTypeDocker))
	is.False(d.Handles(ImageTypeQCOW))
	is.False(d.Handles("anything"))
}

func TestDockerDriver_Run(t *testing.T) {
	d, err := Lookup("docker")
	is := assert.New(t)
	is.NoError(err)
	is.NotNil(d)

	op := &Operation{
		Installation: "test",
		Image:        "test:1.2.3",
		ImageType:    "oci",
	}
	is.NoError(d.Run(op))
}
