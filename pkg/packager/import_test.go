package packager

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/cnab-go/bundle/loader"
	"github.com/stretchr/testify/assert"
)

func TestImport(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "duffle-import-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	is := assert.New(t)

	im := Importer{
		Source:      "testdata/examplebun-0.1.0.tgz",
		Destination: tempDir,
		Loader:      loader.NewLoader(),
	}

	if err := im.Import(); err != nil {
		t.Fatalf("import failed: %v", err)
	}

	expectedBundlePath := filepath.Join(tempDir, "examplebun-0.1.0")
	is.DirExistsf(expectedBundlePath, "expected examplebun to exist")
}

func TestMalformedImport(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "duffle-import-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	im := Importer{
		Source:      "testdata/malformed-0.1.0.tgz",
		Destination: tempDir,
		Loader:      loader.NewLoader(),
	}

	if err = im.Import(); err == nil {
		t.Error("expected malformed bundle error")
	}
}
