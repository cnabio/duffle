package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/stretchr/testify/assert"

	"github.com/deis/duffle/pkg/duffle/home"
)

func TestGetBundleFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	duffleHome = filepath.Join(cwd, "..", "..", "tests", "testdata", "home")
	testHome := home.Home(duffleHome)

	filePath, err := getBundleFile("https://hub.cnlabs.io/helloazure:0.1.0")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(filePath)

	expectedFilepath := filepath.Join(testHome.Cache(), "helloazure-0.1.0.json")

	if filePath != expectedFilepath {
		t.Errorf("got '%v', wanted '%v'", filePath, expectedFilepath)
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
