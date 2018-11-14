package repo

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadIndexReader(t *testing.T) {
	buf := bytes.NewBufferString(`{
	"hub.cnlabs.io/helloworld": {
		"1.0.0": "abcdefghijklmnop",
		"2.0.0": "abcdefghijklmnop"
	},
	"hub.cnlabs.io/goodbyeworld": {
		"1.0.0": "abcdefghijklmnop",
		"2.0.0": "abcdefghijklmnop"
	}
}`)

	l, err := LoadIndexReader(buf)
	if err != nil {
		t.Error(err)
	}

	expectedList := Index{
		"hub.cnlabs.io/helloworld": {
			"1.0.0": "abcdefghijklmnop",
			"2.0.0": "abcdefghijklmnop",
		},
		"hub.cnlabs.io/goodbyeworld": {
			"1.0.0": "abcdefghijklmnop",
			"2.0.0": "abcdefghijklmnop",
		},
	}

	if !reflect.DeepEqual(expectedList, l) {
		t.Errorf("expected lists to be equal, got '%v'", l)
	}

	assert.True(t, l.Delete("hub.cnlabs.io/goodbyeworld"))
	assert.False(t, l.Has("hub.cnlabs.io/goodbyeworld", "1.0.0"))
}
