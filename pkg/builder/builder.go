package builder

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/duffle/manifest"
)

// BundleBuilder defines how a bundle is built and pushed using the supplied app context.
type BundleBuilder interface {
	PrepareBuild(bldr *Builder, appDir string) (*AppContext, *bundle.Bundle, error)
	Build(context.Context, *AppContext) chan *Summary
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
		ID: getulid(),
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

// Summarize returns a function closure that wraps writing SummaryStatusCode.
func Summarize(id, desc string, out chan<- *Summary) func(string, SummaryStatusCode) {
	return func(info string, code SummaryStatusCode) {
		out <- &Summary{StageDesc: desc, StatusText: info, StatusCode: code, BuildID: id}
	}
}

// Complete marks the end of a duffle build stage.
func Complete(id, desc string, out chan<- *Summary, err *error) {
	switch fn := Summarize(id, desc, out); {
	case *err != nil:
		fn(fmt.Sprintf("failure: %v", *err), SummaryFailure)
	default:
		fn("success", SummarySuccess)
	}
}
