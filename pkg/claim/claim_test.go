package claim

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Make sure that the default Result has status and action set.
	claim := New("my_claim")

	assert.Equal(t, "my_claim", claim.Name, "Name is set")
	assert.Equal(t, "unknown", claim.Result.Status)
	assert.Equal(t, "unknown", claim.Result.Action)
}

func TestUpdate(t *testing.T) {
	claim := New("claim")
	oldMod := claim.Modified
	oldUlid := claim.Revision

	time.Sleep(1 * time.Millisecond) // Force the Update to happen at a new time. For those of us who remembered to press the Turbo button.

	claim.Update(ActionInstall, StatusSuccess)

	is := assert.New(t)
	is.NotEqual(oldMod, claim.Modified)
	is.NotEqual(oldUlid, claim.Revision)
	is.Equal("install", claim.Result.Action)
	is.Equal("success", claim.Result.Status)
}
