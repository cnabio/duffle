package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/scothis/ruffle/pkg/version"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	showVersion(buf)
	assert.Equal(t, version.Version, strings.TrimSpace(buf.String()))
}
