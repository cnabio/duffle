package action

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/claim"
	"github.com/deislabs/duffle/pkg/driver"

	"github.com/stretchr/testify/assert"
)

func TestStatus_Run(t *testing.T) {
	out := ioutil.Discard
	b := mockBundle()
	b.Actions["io.cnab.status"] = bundle.Action{}
	st := &Status{Driver: &driver.DebugDriver{}}
	c := &claim.Claim{
		Created:    time.Time{},
		Modified:   time.Time{},
		Name:       "name",
		Revision:   "revision",
		Parameters: map[string]interface{}{},
		Bundle:     b,
	}

	if err := st.Run(c, mockSet, out); err != nil {
		t.Fatal(err)
	}

	st = &Status{Driver: &mockFailingDriver{}}
	assert.Error(t, st.Run(c, mockSet, out))
}
