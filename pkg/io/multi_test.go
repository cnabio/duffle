package io

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiReadCloser(t *testing.T) {
	first := bufCloser("first")
	second := bufCloser("second")
	third := bufCloser("third")
	mrc := MultiReadCloser(first, second, third)

	out, err := ioutil.ReadAll(mrc)
	assert.NoError(t, err)
	assert.Equal(t, "firstsecondthird", string(out))
}

func bufCloser(data string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBufferString(data))
}
