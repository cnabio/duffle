package claim

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Make sure that the default Result has status and action set.
	claim, err := New("my_claim")
	assert.NoError(t, err)

	assert.Equal(t, "my_claim", claim.Name, "Name is set")
	assert.Equal(t, "unknown", claim.Result.Status)
	assert.Equal(t, "unknown", claim.Result.Action)
}

func TestUpdate(t *testing.T) {
	claim, err := New("claim")
	assert.NoError(t, err)
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

func TestValidName(t *testing.T) {
	is := assert.New(t)

	for name, expect := range map[string]bool{
		"M4cb3th":               true,
		"Lady MacBeth":          false, //spaces illegal
		"3_Witches":             true,
		"Banquø":                false, // We could probably loosen this one up
		"King-Duncan":           true,
		"MacDuff@geocities.com": false,
		"hecate":                true, // I wouldn't dare cross Hecate.
	} {
		is.Equal(expect, ValidName.MatchString(name))
	}
}
