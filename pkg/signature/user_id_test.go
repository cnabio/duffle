package signature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUserID(t *testing.T) {
	is := assert.New(t)
	for _, tt := range []struct {
		in     string
		expect UserID
		fail   bool
	}{
		{
			in:     "NAME (COMMENT) <EMAIL@ADDRESS>",
			expect: UserID{Name: "NAME", Comment: "COMMENT", Email: "EMAIL@ADDRESS"},
		},
		{
			in:     "NAME <EMAIL@ADDRESS>",
			expect: UserID{Name: "NAME", Email: "EMAIL@ADDRESS"},
		},
		{
			in:     "EMAIL@ADDRESS",
			expect: UserID{Name: "EMAIL@ADDRESS", Email: "EMAIL@ADDRESS"},
		},
		{
			in:   "<EMAIL@ADDRESS>",
			fail: true,
		},
		{
			in:     "This is a long name (this is a long comment) <this.is+email@some.long.name.com>",
			expect: UserID{Name: "This is a long name", Comment: "this is a long comment", Email: "this.is+email@some.long.name.com"},
		},
		{
			in:     "me () <me@ts>",
			expect: UserID{Name: "me", Comment: "", Email: "me@ts"},
		},
		{
			in:   "me () me@ts",
			fail: true,
		},
		{
			// So it's unclear what we want to do in this case. We can either allow '()' in names or
			// we can read this as "blank name and comment". For now, we'll opt for the former.
			in:     "(foo) <email@example>",
			expect: UserID{Name: "(foo)", Email: "email@example"},
		},
	} {
		res, err := ParseUserID(tt.in)
		if err != nil {
			if tt.fail {
				continue
			}
			t.Errorf("failed on %s: %s", tt.in, err)
			continue
		}

		if tt.fail {
			t.Errorf("expected %s to fail", tt.in)
		}
		is.Equal(res, tt.expect, "failed on %s", tt.in)
	}
}

func TestUserID_String(t *testing.T) {
	u := UserID{
		Name:    "Ahab",
		Comment: "Captain",
		Email:   "ahab@example.com",
	}
	assert.Equal(t, u.String(), "Ahab (Captain) <ahab@example.com>")

	u = UserID{
		Name:  "Captain Ahab",
		Email: "cptahab@hotmail.com",
	}
	assert.Equal(t, u.String(), "Captain Ahab <cptahab@hotmail.com>")
}
