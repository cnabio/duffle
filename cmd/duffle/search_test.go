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
	duffleHome = filepath.Join(cwd, "testdata", "home")

	expectedBundleList := []string{
		"github.com/customorg/duffle-bundles/hello",
		"hello",
	}

	bundleList := search([]string{})
	if !reflect.DeepEqual(bundleList, expectedBundleList) {
		t.Errorf("expected bundle lists to be equal; got '%v', wanted '%v'", bundleList, expectedBundleList)
	}
}
