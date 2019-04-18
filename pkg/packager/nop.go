package packager

import (
	"io"
)

// nop is an ImageStore which does not store images. It is used to construct thin bundles.
type nop struct{}

func newNop() nop {
	return nop{}
}

func (t nop) configure(archiveDir string, logs io.Writer) error {
	return nil
}

func (t nop) add(im string) (string, error) {
	return "", nil
}
