package action

import (
	"testing"
	"time"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/driver"

	"github.com/stretchr/testify/assert"
)

func TestStatus_Run(t *testing.T) {
	st := &Status{Driver: &driver.DebugDriver{}}
	c := &claim.Claim{
		Created:    time.Time{},
		Modified:   time.Time{},
		Name:       "name",
		Revision:   "revision",
		Bundle:     mockBundle(),
		Parameters: map[string]interface{}{},
	}

	if err := st.Run(c, mockSet); err != nil {
		t.Fatal(err)
	}

	st = &Status{Driver: &mockFailingDriver{}}
	assert.Error(t, st.Run(c, mockSet))
}
