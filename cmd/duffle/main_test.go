package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPathy(t *testing.T) {
	is := assert.New(t)
	for path, expect := range map[string]bool{
		"foo": false,
		filepath.Join("this", "is", "a", "path"): true,
		"foo.yaml":                      false,
		filepath.Join("..", "foo.yaml"): true,
	} {
		is.Equal(expect, isPathy(path), "Expected %t, for %s", expect, path)
	}
}
