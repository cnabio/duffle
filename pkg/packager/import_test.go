package packager

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deis/duffle/pkg/loader"

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
		Loader:      loader.NewDetectingLoader(),
	}

	if err := im.Import(); err != nil {
		t.Fatalf("import failed: %v", err)
	}

	expectedBundlePath := filepath.Join(tempDir, "examplebun")
	is.DirExists(expectedBundlePath, "expected examplebun to exist")

	im = Importer{
		Source:      "testdata/malformed-0.1.0.tgz",
		Destination: tempDir,
		Loader:      loader.NewDetectingLoader(),
	}
	is.Error(im.Import(), "expected malformed bundle error")
}
