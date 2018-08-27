package action

import (
	"testing"
	"time"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/driver"

	"github.com/stretchr/testify/assert"
)

func TestUninstall_Run(t *testing.T) {
	c := &claim.Claim{
		Created:    time.Time{},
		Modified:   time.Time{},
		Name:       "name",
		Revision:   "revision",
		Bundle:     "fake/bundle:0.1.0",
		Parameters: map[string]interface{}{},
	}

	uninst := &Uninstall{Driver: &driver.DebugDriver{}}
	assert.NoError(t, uninst.Run(c, mockSet))
	if c.Created == c.Modified {
		t.Error("Claim was not updated with modified time stamp during uninstallafter uninstall action")
	}

	if c.Result.Action != claim.ActionUninstall {
		t.Errorf("Claim result action not successfully updated. Expected %v, got %v", claim.ActionUninstall, c.Result.Action)
	}
	if c.Result.Status != claim.StatusSuccess {
		t.Errorf("Claim result status not successfully updated. Expected %v, got %v", claim.StatusSuccess, c.Result.Status)
	}

	uninst = &Uninstall{Driver: &mockFailingDriver{}}
	assert.Error(t, uninst.Run(c, mockSet))

	uninst = &Uninstall{Driver: &mockFailingDriver{shouldHandle: true}}
	assert.Error(t, uninst.Run(c, mockSet))
	if c.Result.Message == "" {
		t.Error("Expected error message in claim result message")
	}

	if c.Result.Action != claim.ActionUninstall {
		t.Errorf("Expected claim result action to be %v, got %v", claim.ActionUninstall, c.Result.Action)
	}

	if c.Result.Status != claim.StatusFailure {
		t.Errorf("Expected claim result status to be %v, got %v", claim.StatusFailure, c.Result.Status)
	}
}
