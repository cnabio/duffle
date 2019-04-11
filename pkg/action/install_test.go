package action

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/deislabs/duffle/pkg/claim"
	"github.com/deislabs/duffle/pkg/driver"

	"github.com/stretchr/testify/assert"
)

// makes sure Install implements Action interface
var _ Action = &Install{}

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
