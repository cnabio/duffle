package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalLoader(t *testing.T) {
	is := assert.New(t)

	l := LocalLoader{source: "../../tests/testdata/bundles/foo.json"}
	bundle, err := l.Load()
	if err != nil {
		t.Fatalf("cannot load bundle: %v", err)
	}

	is.Equal("foo", bundle.Name)
	is.Equal("1.0", bundle.Version)
}
