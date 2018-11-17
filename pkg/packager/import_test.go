package packager

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestImport(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "duffle-import-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	im := Importer{
		Source:      "testdata/examplebun-0.1.0.tgz",
		Destination: tempDir,
	}

	if err := im.Import(); err != nil {
		t.Errorf("Expected no error, got error: %v", err)
	}

	expectedBundlePath := filepath.Join(tempDir, "examplebun")
	fi, err := os.Stat(expectedBundlePath)
	if err != nil {
		t.Errorf("Expected no error examining decompressed bundle but go error %s", err)
	} else if !fi.IsDir() {
		t.Errorf("Expected %s to be directory but is not", expectedBundlePath)
	}

	im = Importer{
		Source:      "testdata/malformed-0.1.0.tgz",
		Destination: tempDir,
	}
	if err := im.Import(); err == nil {
		t.Error("Expected error due to malformed bundle but got none")
	}
}
