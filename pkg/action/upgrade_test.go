package action

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/deislabs/duffle/pkg/claim"
	"github.com/deislabs/duffle/pkg/driver"

	"github.com/stretchr/testify/assert"
)

func TestUpgrade_Run(t *testing.T) {
	out := ioutil.Discard

	c := &claim.Claim{
		Created:    time.Time{},
		Modified:   time.Time{},
		Name:       "name",
		Revision:   "revision",
		Bundle:     mockBundle(),
		Parameters: map[string]interface{}{},
	}

	upgr := &Upgrade{Driver: &driver.DebugDriver{}}
	assert.NoError(t, upgr.Run(c, mockSet, out))
	if c.Created == c.Modified {
		t.Error("Claim was not updated with modified time stamp during upgrade action")
	}

	if c.Result.Action != claim.ActionUpgrade {
		t.Errorf("Claim result action not successfully updated. Expected %v, got %v", claim.ActionUninstall, c.Result.Action)
	}
	if c.Result.Status != claim.StatusSuccess {
		t.Errorf("Claim result status not successfully updated. Expected %v, got %v", claim.StatusSuccess, c.Result.Status)
	}

	upgr = &Upgrade{Driver: &mockFailingDriver{}}
	assert.Error(t, upgr.Run(c, mockSet, out))

	upgr = &Upgrade{Driver: &mockFailingDriver{shouldHandle: true}}
	assert.Error(t, upgr.Run(c, mockSet, out))
	if c.Result.Message == "" {
		t.Error("Expected error message in claim result message")
	}

	if c.Result.Action != claim.ActionUpgrade {
		t.Errorf("Expected claim result action to be %v, got %v", claim.ActionUpgrade, c.Result.Action)
	}

	if c.Result.Status != claim.StatusFailure {
		t.Errorf("Expected claim result status to be %v, got %v", claim.StatusFailure, c.Result.Status)
	}
}
