package docker

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/cnabio/duffle/pkg/imagebuilder"
)

func TestArchiveSrc(t *testing.T) {
	c := &Builder{}
	err := archiveSrc(filepath.Join("..", "..", "..", "tests", "testdata", "builder", "simple"), c)
	if err != nil {
		t.Error(err)
	}
	defer c.BuildContext.Close()

	buf, err := ioutil.ReadAll(c.BuildContext)
	if err != nil {
		t.Error(err)
	}

	// testdata/simple should have a length of 2048 bytes
	if len(buf) != 2048 {
		t.Errorf("expected non-zero archive length, got %d", len(buf))
	}
}

// tests that Builder implements imagebuilder.ImageBuilder interface
func TestBuilder_implBuilder(t *testing.T) {
	var _ imagebuilder.ImageBuilder = (*Builder)(nil)
}
