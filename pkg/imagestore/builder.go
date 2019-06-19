package imagestore

import (
	"io"
)

// Builder is a means of creating image stores.
type Builder interface {
	// ArchiveDir creates a fresh Builder with the given archive directory.
	ArchiveDir(string) Builder

	// Logs creates a fresh builder with the given log stream.
	Logs(io.Writer) Builder

	// Build creates an image store.
	Build() (Store, error)
}
