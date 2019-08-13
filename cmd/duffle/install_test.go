package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/cnab-go/bundle"

	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"

	"github.com/deislabs/duffle/pkg/duffle/home"
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
			bundle, err := getBundleFilepath(tc.File, duffleHome)
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
	defs := definition.Definitions{
		"first":  {Type: "string"},
		"second": {Type: "boolean"},
		"third":  {Type: "integer"},
	}
	params := map[string]bundle.Parameter{
		"str":    {Definition: "first"},
		"second": {Definition: "second"},
		"int":    {Definition: "third"},
		"bad":    {Definition: "three hundred and thirty third"},
	}

	setVals := []string{"str=foo", "second=true", "int=2", "fourth"}
	o, err := overrides(setVals, params, defs)
	is.NoError(err)

	is.Len(o, 3)
	is.Equal(o["str"].(string), "foo")
	is.True(o["second"].(bool))
	is.Equal(o["int"].(int), 2)

	// We expect an error if we pass a param that was not defined:
	_, err = overrides([]string{"undefined=foo"}, params, defs)
	is.Error(err)

	// We expect an error if we pass a param that was not defined even if the name matches a definition:
	_, err = overrides([]string{"first=foo"}, params, defs)
	is.Error(err)

	// We expect an error if we pass a param whose definition does not exist:
	_, err = overrides([]string{"bad=worse"}, params, defs)
	is.Error(err)
}
