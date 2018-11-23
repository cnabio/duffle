package repo

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/stretchr/testify/assert"
)

func namedOrDie(v string) reference.Named {
	res, err := reference.ParseNormalizedNamed(v)
	if err != nil {
		panic(err)
	}
	return res
}

func taggedOrDie(name, tag string) reference.NamedTagged {
	res, err := reference.WithTag(namedOrDie(name), tag)
	if err != nil {
		panic(err)
	}
	return res
}

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

	is := assert.New(t)
	revs, ok := l.GetVersions(namedOrDie("hub.cnlabs.io/goodbyeworld"))
	is.True(ok)
	is.Len(revs, 2)
	is.Equal("abcdefghijklmnop", revs["1.0.0"])

	is.True(l.DeleteAll(namedOrDie("hub.cnlabs.io/goodbyeworld")))
	v1, err := reference.WithTag(namedOrDie("hub.cnlabs.io/goodbyeworld"), "1.0.0")
	is.NoError(err)
	is.False(l.Has(v1))
	is.False(l.DeleteVersion(taggedOrDie("nosuchname", "0.1.2")))

	is.True(l.DeleteVersion(taggedOrDie("hub.cnlabs.io/helloworld", "2.0.0")))
	is.True(l.Has(taggedOrDie("hub.cnlabs.io/helloworld", "1.0.0")))
	is.False(l.Has(taggedOrDie("hub.cnlabs.io/helloworld", "2.0.0")))
}
