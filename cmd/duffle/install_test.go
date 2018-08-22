package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/deis/duffle/pkg/duffle/home"
)

func TestGetBundleFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	duffleHome = filepath.Join(cwd, "..", "..", "tests", "testdata", "home")
	testHome := home.Home(duffleHome)

	filePath, repo, err := getBundleFile("foo")
	if err != nil {
		t.Error(err)
	}

	expectedFilepath := filepath.Join(testHome.Repositories(), testHome.DefaultRepository(), "bundles", "foo.json")
	expectedRepo := testHome.DefaultRepository()

	if filePath != expectedFilepath {
		t.Errorf("got '%v', wanted '%v'", filePath, expectedFilepath)
	}

	if repo != expectedRepo {
		t.Errorf("got '%v', wanted '%v'", repo, expectedRepo)
	}
}
