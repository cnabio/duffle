package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPathy(t *testing.T) {
	is := assert.New(t)
	thispath := filepath.Join("this", "is", "a", "path")
	fooya := filepath.Join("..", "foo.yaml")
	for path, expect := range map[string]bool{
		"foo":      false,
		thispath:   true,
		"foo.yaml": false,
		fooya:      true,
	} {
		is.Equal(expect, isPathy(path), "Expected %t, for %s", expect, path)
	}
}
