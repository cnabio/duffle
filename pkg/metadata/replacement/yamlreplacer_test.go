package replacement

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanReplace(t *testing.T) {
	source := "a: 1\nb:\n  c: d\n  e: f"
	r := NewYAMLReplacer()
	result, err := r.Replace(source, "b.c", "test")
	if err != nil {
		t.Fatalf("Replace failed: %s", err)
	}

	expected := strings.Replace(source, "d", "test", -1)

	is := assert.New(t)
	is.Equal(strings.TrimSpace(expected), strings.TrimSpace(result))
}
