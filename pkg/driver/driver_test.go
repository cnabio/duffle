package driver

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ Driver = &DockerDriver{}
var _ Driver = &DebugDriver{}

func TestLookup_UnsupportedDriver(t *testing.T) {
	d, err := Lookup("no_such_driver")

	assert.Nil(t, d)
	assert.Error(t, err)
}

func TestLookup_CommandDriver(t *testing.T) {
	d, err := Lookup("command")

	assert.NoError(t, err)
	assert.IsType(t, d, &CommandDriver{})
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
		Out:          ioutil.Discard,
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

	// Don't actually run Docker
	d.(*DockerDriver).Simulate = true

	is := assert.New(t)
	is.NoError(err)
	is.NotNil(d)

	op := &Operation{
		Installation: "test",
		Image:        "test:1.2.3",
		ImageType:    "oci",
		Environment:  map[string]string{},
		Files:        map[string]string{},
		Out:          ioutil.Discard,
	}
	is.NoError(d.Run(op))
}

func TestACIDriver_Handles(t *testing.T) {
	d, err := Lookup("docker")
	is := assert.New(t)
	is.NoError(err)
	is.NotNil(d)
	is.True(d.Handles(ImageTypeDocker))
	is.True(d.Handles(ImageTypeOCI))
	is.False(d.Handles(ImageTypeQCOW))
	is.False(d.Handles("anything"))
}

func TestACIDriver_Run(t *testing.T) {
	d, err := Lookup("aci")

	// Don't actually run Docker
	d.(*ACIDriver).Simulate = true

	is := assert.New(t)
	is.NoError(err)
	is.NotNil(d)

	op := &Operation{
		Installation: "test",
		Image:        "test:1.2.3",
		ImageType:    "oci",
		Environment:  map[string]string{},
		Files:        map[string]string{},
		Out:          ioutil.Discard,
	}
	is.NoError(d.Run(op))
}
