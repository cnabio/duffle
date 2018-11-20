package packager

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestExport(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "duffle-export-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	source, err := filepath.Abs(filepath.Join("testdata", "examplebun"))
	if err != nil {
		t.Fatal(err)
	}
	ex := Exporter{
		Source: source,
		Full:   false,
	}

	tempPWD, err := ioutil.TempDir("", "duffle-export-test")
	if err != nil {
		t.Fatal(err)
	}
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(pwd)
	if err := os.Chdir(tempPWD); err != nil {
		t.Fatal(err)
	}

	if err := ex.Export(); err != nil {
		t.Errorf("Expected no error, got error: %v", err)
	}

	expectedFile := "examplebun-0.1.0.tgz"
	_, err = os.Stat(expectedFile)
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but was not created", expectedFile)
	} else if err != nil {
		t.Errorf("Error with compressed bundle file: %v", err)
	}
}

func TestExportCreatesFileProperly(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "duffle-export-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	ex := Exporter{
		Source:      "testdata/examplebun",
		Destination: filepath.Join(tempDir, "random-directory", "examplebun-whatev.tgz"),
		Full:        false,
	}

	if err := ex.Export(); err == nil {
		t.Error("Expected path does not exist error, got no error")
	}

	if err := os.MkdirAll(filepath.Join(tempDir, "random-directory"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := ex.Export(); err != nil {
		t.Errorf("Expected no error, got error: %s", err)
	}

	expectedFile := filepath.Join(tempDir, "random-directory", "examplebun-whatev.tgz")
	_, err = os.Stat(expectedFile)
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but was not created", expectedFile)
	} else if err != nil {
		t.Errorf("Error with compressed bundle archive: %v", err)
	}
}
