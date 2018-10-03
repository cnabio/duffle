package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUlid(t *testing.T) {
	// Basically, we're checking that the channel is getting replenished with each call.
	first := getulid()
	second := getulid()
	assert.NotEqual(t, first, second)
}
