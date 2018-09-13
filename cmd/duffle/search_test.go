package main

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSearch(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	duffleHome = filepath.Join(cwd, "..", "..", "tests", "testdata", "home")

	if _, err := search([]string{}); err != nil {
		t.Error(err)
	}

	expectedBundleList := []string{
		"helloazure",
		"helloworld",
	}

	bundleList, err := search([]string{"hello"})
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(bundleList, expectedBundleList) {
		t.Errorf("expected bundle lists to be equal; got '%v', wanted '%v'", bundleList, expectedBundleList)
	}
}
