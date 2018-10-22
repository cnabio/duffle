package main

import (
	"log"
	"os"
	"path/filepath"
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

	bundleList, err := search([]string{"hello"})
	if err != nil {
		t.Error(err)
	}

	if len(bundleList) == 0 {
		t.Error("expected to find at least one bundle with the keyword 'hello'")
	}
}
