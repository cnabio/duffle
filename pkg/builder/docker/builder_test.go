package docker

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestArchiveSrc(t *testing.T) {
	ctx, err := archiveSrc(filepath.Join("..", "..", "..", "tests", "testdata", "builder", "simple"), "")
	if err != nil {
		t.Error(err)
	}
	defer ctx.BuildContext.Close()

	buf, err := ioutil.ReadAll(ctx.BuildContext)
	if err != nil {
		t.Error(err)
	}

	// testdata/simple should have a length of 2048 bytes
	if len(buf) != 2048 {
		t.Errorf("expected non-zero archive length, got %d", len(buf))
	}
}
