package crud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDBName(t *testing.T) {
	testURL := "mongodb://localhost:1234/foo"
	name, err := parseDBName(testURL)
	assert.NoError(t, err)
	assert.Equal(t, "foo", name)

	name, err = parseDBName("mongodb://localhost:1234")
	assert.NoError(t, err)
	assert.Equal(t, "", name)

	name, err = parseDBName("://example.com:8080")
	assert.Error(t, err)
	assert.Equal(t, "", name)
}
