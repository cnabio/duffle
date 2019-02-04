package action

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/scothis/ruffle/pkg/bundle"
	"github.com/scothis/ruffle/pkg/claim"
	"github.com/scothis/ruffle/pkg/driver"

	"github.com/stretchr/testify/assert"
)

func TestRunCustom(t *testing.T) {
	out := ioutil.Discard
	is := assert.New(t)

	rc := &RunCustom{
		Driver: &driver.DebugDriver{},
		Action: "test",
	}
	c := &claim.Claim{
		Created:    time.Time{},
		Modified:   time.Time{},
		Name:       "runcustom",
		Revision:   "revision",
		Bundle:     mockBundle(),
		Parameters: map[string]interface{}{},
	}

	if err := rc.Run(c, mockSet, out); err != nil {
		t.Fatal(err)
	}
	is.Equal(claim.StatusSuccess, c.Result.Status)
	is.Equal("test", c.Result.Action)

	// Make sure we don't allow forbidden custom actions
	rc.Action = "install"
	is.Error(rc.Run(c, mockSet, out))

	// Get rid of custom actions, and this should fail
	rc.Action = "test"
	c.Bundle.Actions = map[string]bundle.Action{}
	if err := rc.Run(c, mockSet, out); err == nil {
		t.Fatal("Unknown action should fail")
	}
}
