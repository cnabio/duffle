package action

import (
	"testing"
	"time"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/driver"

	"github.com/stretchr/testify/assert"
)

func TestInstall_Run(t *testing.T) {
	c := &claim.Claim{
		Created:    time.Time{},
		Modified:   time.Time{},
		Name:       "name",
		Revision:   "revision",
		Bundle:     mockBundle(),
		Parameters: map[string]interface{}{},
	}

	inst := &Install{Driver: &driver.DebugDriver{}}
	assert.NoError(t, inst.Run(c, mockSet))

	inst = &Install{Driver: &mockFailingDriver{}}
	assert.Error(t, inst.Run(c, mockSet))

	inst = &Install{Driver: &mockFailingDriver{shouldHandle: true}}
	assert.Error(t, inst.Run(c, mockSet))
}
