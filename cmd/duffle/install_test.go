package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/duffle/home"
)

func TestGetBundleFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	duffleHome = filepath.Join(cwd, "..", "..", "tests", "testdata", "home")
	testHome := home.Home(duffleHome)

	tests := []struct {
		Name             string
		File             string
		ExpectedFilepath string
	}{
		{
			Name:             "helloazure",
			File:             "https://hub.cnlabs.io/helloazure:0.1.0",
			ExpectedFilepath: filepath.Join(testHome.Cache(), "helloazure-0.1.0.json"),
		},
		{
			Name:             "namespaced helloazure",
			File:             "https://hub.cnlabs.io/library/helloazure:0.1.0",
			ExpectedFilepath: filepath.Join(testHome.Cache(), "helloazure-0.1.0.json"),
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			filePath, err := getBundleFile(tc.File)
			if err != nil {
				t.Error(err)
			}
			defer os.Remove(filePath)

			if filePath != tc.ExpectedFilepath {
				t.Errorf("got '%v', wanted '%v'", filePath, tc.ExpectedFilepath)
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
