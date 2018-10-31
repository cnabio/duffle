package builder

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/duffle/manifest"
)

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
	bf := &bundle.Bundle{
		Name:        ctx.Manifest.Name,
		Description: ctx.Manifest.Description,
		Images:      []bundle.Image{},
		Keywords:    ctx.Manifest.Keywords,
		Maintainers: ctx.Manifest.Maintainers,
		Parameters:  ctx.Manifest.Parameters,
		Credentials: ctx.Manifest.Credentials,
	}

	for _, c := range ctx.Components {
		if err := c.PrepareBuild(ctx); err != nil {
			return nil, nil, err
		}

		if c.Name() == "cnab" {
			bf.InvocationImages = []bundle.InvocationImage{
				{
					Image:     c.URI(),
					ImageType: c.Type(),
				}}
			bf.Version = strings.Split(c.URI(), ":")[1]
		} else {
			bf.Images = append(bf.Images, bundle.Image{Name: c.Name(), URI: c.URI()})
		}
	}

	app := &AppContext{
		ID:   bldr.ID,
		Bldr: bldr,
		Ctx:  ctx,
		Log:  os.Stdout,
	}

	return app, bf, nil
}

// Build passes the context of each component to its respective builder
func (b *Builder) Build(ctx context.Context, app *AppContext) error {
	if err := buildComponents(ctx, app); err != nil {
		return fmt.Errorf("error building components: %v", err)
	}
	return nil
}

func buildComponents(ctx context.Context, app *AppContext) (err error) {
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
			time.Sleep(time.Second)
		}
	}
	return nil
}
