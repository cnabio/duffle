package repo

import (
	"bytes"
	"reflect"
	"sort"
	"testing"

	"github.com/Masterminds/semver"

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

	is := assert.New(t)
	revs, ok := l.GetVersions("hub.cnlabs.io/goodbyeworld")
	is.True(ok)
	is.Len(revs, 2)
	is.Equal("abcdefghijklmnop", revs[0].Digest)

	is.True(l.Delete("hub.cnlabs.io/goodbyeworld"))
	is.False(l.Has("hub.cnlabs.io/goodbyeworld", "1.0.0"))

	is.False(l.DeleteVersion("nosuchname", "0.1.2"))
	is.True(l.DeleteVersion("hub.cnlabs.io/helloworld", "2.0.0"))
	is.True(l.Has("hub.cnlabs.io/helloworld", "1.0.0"))
	is.False(l.Has("hub.cnlabs.io/helloworld", "2.0.0"))
}

func TestBundleVersionSortByVersion(t *testing.T) {
	byVersion := ByVersion{
		BundleVersion{
			Version: semver.MustParse("0.1.0"),
		},
		BundleVersion{
			Version: semver.MustParse("0.2.0"),
		},
	}

	sort.Sort(byVersion)
	if byVersion[0].Version.String() != "0.1.0" {
		t.Errorf("expected 0.1.0, got %s", byVersion[0].Version.String())
	}

	sort.Sort(sort.Reverse(byVersion))
	if byVersion[0].Version.String() != "0.2.0" {
		t.Errorf("expected 0.2.0, got %s", byVersion[0].Version.String())
	}
}
