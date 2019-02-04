package action

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/scothis/ruffle/pkg/claim"
	"github.com/scothis/ruffle/pkg/driver"

	"github.com/stretchr/testify/assert"
)

func TestInstall_Run(t *testing.T) {
	out := ioutil.Discard

	c := &claim.Claim{
		Created:    time.Time{},
		Modified:   time.Time{},
		Name:       "name",
		Revision:   "revision",
		Bundle:     mockBundle(),
		Parameters: map[string]interface{}{},
	}

	inst := &Install{Driver: &driver.DebugDriver{}}
	assert.NoError(t, inst.Run(c, mockSet, out))

	inst = &Install{Driver: &mockFailingDriver{}}
	assert.Error(t, inst.Run(c, mockSet, out))

	inst = &Install{Driver: &mockFailingDriver{shouldHandle: true}}
	assert.Error(t, inst.Run(c, mockSet, out))
}
