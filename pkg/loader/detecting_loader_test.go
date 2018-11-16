package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectingLoader_Signed(t *testing.T) {
	is := assert.New(t)

	b, err := NewDetectingLoader().Load(testSigned)
	is.NoError(err)
	is.Equal(b.Name, "example_bundle")
}

func TestDetectingLoader_Unsigned(t *testing.T) {
	is := assert.New(t)

	bundle, err := NewDetectingLoader().Load(testFooJSON)
	if err != nil {
		t.Fatalf("cannot load bundle: %v", err)
	}

	is.Equal("foo", bundle.Name)
	is.Equal("1.0", bundle.Version)
}
