package replacement

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanReplaceInJSON(t *testing.T) {
	source := `{
	"a": 1,
	"b": {
		"c": "d",
		"e": "f"
	}
}`
	r := NewJSONReplacer("\t")
	result, err := r.Replace(source, "b.c", "test")
	if err != nil {
		t.Fatalf("Replace failed: %s", err)
	}

	expected := strings.Replace(source, "d", "test", -1)

	is := assert.New(t)
	is.Equal(strings.TrimSpace(expected), strings.TrimSpace(result))
}

func TestJSONErrorIfPathNotFound(t *testing.T) {
	source := `{
	"a": 1,
	"b": {
		"c": "d",
		"e": "f"
	}
}`
	r := NewJSONReplacer("\t")

	_, err := r.Replace(source, "b.c.d", "test")
	if err != ErrSelectorNotFound {
		t.Error("Expected path not found error for b.c.d")
	}

	_, err = r.Replace(source, "b.d", "test")
	if err != ErrSelectorNotFound {
		t.Error("Expected path not found error for b.d")
	}
}
