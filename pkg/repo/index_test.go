package repo

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/Masterminds/semver"

	"github.com/stretchr/testify/assert"
)

func TestLoadIndexReader(t *testing.T) {
	buf := bytes.NewBufferString(`{
	"helloworld": {
		"1.0.0": "abcdefghijklmnop",
		"2.0.0": "abcdefghijklmnop"
	},
	"goodbyeworld": {
		"1.0.0": "abcdefghijklmnop",
		"2.0.0": "abcdefghijklmnop"
	}
}`)

	l, err := LoadIndexReader(buf)
	if err != nil {
		t.Error(err)
	}

	expectedList := Index{
		"helloworld": {
			"1.0.0": "abcdefghijklmnop",
			"2.0.0": "abcdefghijklmnop",
		},
		"goodbyeworld": {
			"1.0.0": "abcdefghijklmnop",
			"2.0.0": "abcdefghijklmnop",
		},
	}

	if !reflect.DeepEqual(expectedList, l) {
		t.Errorf("expected lists to be equal, got '%v'", l)
	}

	is := assert.New(t)
	revs, ok := l.GetVersions("goodbyeworld")
	is.True(ok)
	is.Len(revs, 2)
	is.Equal("abcdefghijklmnop", revs[0].Digest)

	is.True(l.Delete("goodbyeworld"))
	is.False(l.Has("goodbyeworld", "1.0.0"))

	is.False(l.DeleteVersion("nosuchname", "0.1.2"))
	is.True(l.DeleteVersion("helloworld", "2.0.0"))
	is.True(l.Has("helloworld", "1.0.0"))
	is.False(l.Has("helloworld", "2.0.0"))
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

func TestGet(t *testing.T) {
	is := assert.New(t)
	testList := Index{
		"helloworld": {
			"":      "malformed",
			"1.0.0": "abcd",
			"2.0.0": "efgh",
		},
		"goodbyeworld": {
			"1.0.0": "ijkl",
			"2.0.0": "mnop",
		},
	}

	testCases := []struct {
		Name           string
		Version        string
		ExpectedDigest string
		ExpectedErr    error
	}{
		{
			Name:           "helloworld",
			Version:        "",
			ExpectedDigest: "efgh",
			ExpectedErr:    nil,
		},
		{
			Name:           "mnop",
			Version:        "",
			ExpectedDigest: "mnop",
			ExpectedErr:    nil,
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(fmt.Sprintf("name=%s;version=%s", tc.Name, tc.Version), func(t *testing.T) {
			t.Parallel()
			actualDigest, actualErr := testList.Get(tc.Name, tc.Version)
			is.Equal(actualDigest, tc.ExpectedDigest)
			is.Equal(actualErr, tc.ExpectedErr)
		})
	}
}
