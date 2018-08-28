package bbuilder

import (
	"context"
	"io"
	"path/filepath"

	"github.com/deis/duffle/pkg/builder"
	"github.com/deis/duffle/pkg/duffle/manifest"
)

// BundleBuilder defines how a bundle is built and pushed using the supplied app context.
type BundleBuilder interface {
	PrepareBuild(bldr *Builder, appDir string) (*AppContext, error)
	Build(context.Context, *AppContext, chan<- *builder.Summary) error
}

// Builder defines how to interact with a bundle builder
type Builder struct {
	ID            string
	BundleBuilder BundleBuilder
	LogsDir       string
}

// New returns a new Builder
func New() *Builder {
	return &Builder{
		ID: builder.GetUlid(),
	}
}

// Logs returns the path to the build logs.
//
// Set after Up is called (otherwise "").
func (b *Builder) Logs(appName string) string {
	return filepath.Join(b.LogsDir, appName, b.ID)
}

// Context contains information about the application
type Context struct {
	Manifest   manifest.Manifest
	AppDir     string
	Components []Component
}

// Component contains the information of a built component
type Component interface {
	URI() string
	Digest() string
}

// AppContext contains state information carried across various duffle stage boundaries
type AppContext struct {
	Bldr *Builder
	Ctx  *Context
	Log  io.WriteCloser
	ID   string
}
