package imagestoremocks

import (
	"io"

	"github.com/deislabs/duffle/pkg/imagestore"
)

type MockBuilder struct {
	ArchiveDirStub func(string)
	LogsStub       func(io.Writer)
	BuildStub      func() (imagestore.Store, error)
}

func (b *MockBuilder) ArchiveDir(archiveDir string) imagestore.Builder {
	b.ArchiveDirStub(archiveDir)
	return b
}

func (b *MockBuilder) Logs(logs io.Writer) imagestore.Builder {
	b.LogsStub(logs)
	return b
}

func (b *MockBuilder) Build() (imagestore.Store, error) {
	return b.BuildStub()
}
