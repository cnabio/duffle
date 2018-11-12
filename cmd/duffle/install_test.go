package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/reference"
)

func TestGetBundle(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	duffleHome = filepath.Join(cwd, "..", "..", "tests", "testdata", "home")
	testHome := home.Home(duffleHome)

	tests := []struct {
		Name           string
		File           string
		ExpectedDigest string
	}{
		{
			Name:           "helloazure",
			File:           "https://hub.cnlabs.io/helloazure:0.1.0",
			ExpectedDigest: filepath.Join(testHome.Bundles(), "0425467240c734b641673bc2d39433311223ff26"),
		},
		{
			Name:           "namespaced helloazure",
			File:           "https://hub.cnlabs.io/library/helloazure:0.1.0",
			ExpectedDigest: filepath.Join(testHome.Bundles(), "381ba3be9b701ce266bb805f52e3c26a8f8571c6"),
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			bundle, err := pullBundle(tc.File, true)
			if err != nil {
				t.Error(err)
			}
			defer os.Remove(bundle)

			if bundle != tc.ExpectedDigest {
				t.Errorf("got '%v', wanted '%v'", bundle, tc.ExpectedDigest)
			}
		})
	}
}

func TestOverrides(t *testing.T) {
	is := assert.New(t)
	// overrides(overrides []string, paramDefs map[string]bundle.ParameterDefinition)
	defs := map[string]bundle.ParameterDefinition{
		"first":  {DataType: "string"},
		"second": {DataType: "bool"},
		"third":  {DataType: "int"},
	}

	setVals := []string{"first=foo", "second=true", "third=2", "fourth"}
	o, err := overrides(setVals, defs)
	is.NoError(err)

	is.Len(o, 3)
	is.Equal(o["first"].(string), "foo")
	is.True(o["second"].(bool))
	is.Equal(o["third"].(int), 2)

	// We expect an error if we pass a param that was not defined:
	_, err = overrides([]string{"undefined=foo"}, defs)
	is.Error(err)
}

func TestRepoURLFromReference(t *testing.T) {
	tests := []struct {
		Name      string
		Reference reference.NamedTagged
		Expected  string
	}{
		{
			Name: "basic repository",
			// discard the digest; we don't care about that
			Reference: &NamedRepository{"hub.cnlabs.io/helloazure", "0.1.0", ""},
			Expected:  "https://hub.cnlabs.io/repositories/helloazure/tags/0.1.0",
		},
		{
			Name: "namespaced repository",
			// discard the digest; we don't care about that
			Reference: &NamedRepository{"hub.cnlabs.io/library/helloazure", "0.2.0", ""},
			Expected:  "https://hub.cnlabs.io/repositories/library/helloazure/tags/0.2.0",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			url, err := repoURLFromReference(tc.Reference)
			if err != nil {
				t.Error(err)
			}

			if url.String() != tc.Expected {
				t.Errorf("expected '%s', got '%s'", tc.Expected, url.String())
			}
		})
	}
}
