package builder

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/duffle/manifest"
	"github.com/deis/duffle/pkg/osutil"
)

// // ComponentBuilder defines how a bundle is built and pushed using the supplied app context.
// type ComponentBuilder interface {
// 	PrepareBuild(bldr *Builder, mfst *manifest.Manifest, appDir string) (*AppContext, *bundle.Bundle, error)
// 	Build(context.Context, *AppContext, chan *Summary) error
// }

// Builder defines how to interact with a bundle builder
type Builder struct {
	ID      string
	LogsDir string
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
	Manifest   *manifest.Manifest
	AppDir     string
	Components []Component
}

// Component contains the information of a built component
type Component interface {
	Name() string
	Type() string
	URI() string
	Digest() string

	PrepareBuild(*Context) error
	Build(context.Context, *AppContext) error
}

// AppContext contains state information carried across various duffle stage boundaries
type AppContext struct {
	Bldr *Builder
	Ctx  *Context
	Log  io.WriteCloser
	ID   string
}

// PrepareBuild prepares a build
func (b *Builder) PrepareBuild(bldr *Builder, mfst *manifest.Manifest, appDir string, components []Component) (*AppContext, *bundle.Bundle, error) {
	ctx := &Context{
		AppDir:     appDir,
		Components: components,
		Manifest:   mfst,
	}

	if err := osutil.EnsureDirectory(filepath.Dir(bldr.Logs(ctx.Manifest.Name))); err != nil {
		return nil, nil, err
	}

	logf, err := os.OpenFile(bldr.Logs(ctx.Manifest.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, nil, err
	}

	var wg sync.WaitGroup
	wg.Add(len(ctx.Components))
	bf := &bundle.Bundle{Name: ctx.Manifest.Name}

	for _, c := range ctx.Components {
		go func(c Component) {
			defer wg.Done()

			err = c.PrepareBuild(ctx)
			if err != nil {
				fmt.Printf("ERROR: %v", err)
			}

			// TODO - concurrency in appending to a slice in a goroutine?
			if c.Name() == "cnab" {
				bf.InvocationImage = bundle.InvocationImage{
					Image:     c.URI(),
					ImageType: c.Type(),
				}
				bf.Version = strings.Split(c.URI(), ":")[1]
				return
			}
			bf.Images = append(bf.Images, bundle.Image{Name: c.Name(), URI: c.URI()})
		}(c)
	}

	wg.Wait()

	app := &AppContext{
		ID:   bldr.ID,
		Bldr: bldr,
		Ctx:  ctx,
		Log:  logf,
	}

	return app, bf, nil
}

// Build passes the context of each component to its respective builder
func (b *Builder) Build(ctx context.Context, app *AppContext) chan *Summary {
	ch := make(chan *Summary, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func(app *AppContext) {
		defer wg.Done()
		log.SetOutput(app.Log)
		if err := buildComponents(ctx, app, ch); err != nil {
			log.Printf("error building components %v", err)
		}
	}(app)

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func buildComponents(ctx context.Context, app *AppContext, out chan *Summary) (err error) {
	const stageDesc = "Building CNAB components"
	defer Complete(app.ID, stageDesc, out, &err)
	summary := Summarize(app.ID, stageDesc, out)
	summary("started", SummaryOngoing)
	errc := make(chan error)

	go func() {
		defer close(errc)
		var wg sync.WaitGroup
		wg.Add(len(app.Ctx.Components))

		for _, c := range app.Ctx.Components {
			go func(c Component) {
				defer wg.Done()
				err = c.Build(ctx, app)
				if err != nil {
					errc <- fmt.Errorf("error building component %v: %v", c.Name(), err)
				}
			}(c)
		}

		wg.Wait()
	}()

	for errc != nil {
		select {
		case err, ok := <-errc:
			if !ok {
				errc = nil
				continue
			}
			return err
		default:
			summary("ongoing", SummaryOngoing)
			time.Sleep(time.Second)
		}
	}
	return nil
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
