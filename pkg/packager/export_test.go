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

	ex := Exporter{
		Source:      "testdata/examplebun",
		Destination: tempDir,
		Full:        false,
	}

	if err := ex.Export(); err != nil {
		t.Errorf("Expected no error, got error: %v", err)
	}
	expectedFile := filepath.Join(tempDir, "examplebun-0.1.0.tgz")
	_, err = os.Stat(expectedFile)
	if err != nil && os.IsNotExist(err) {
		t.Errorf("Expected %s to exist but was not created", expectedFile)
	} else if err != nil {
		t.Errorf("Error with compressed bundle file: %v", err)
	}
}
