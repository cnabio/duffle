package builder

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/deis/duffle/pkg/duffle"
	"github.com/deis/duffle/pkg/duffle/manifest"
	"github.com/deis/duffle/pkg/osutil"
	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	// DockerignoreFilename is the filename for Docker's ignore file.
	DockerignoreFilename = ".dockerignore"
)

// Builder contains information about the build environment
type Builder struct {
	ID               string
	ContainerBuilder ContainerBuilder
	LogsDir          string
}

// ContainerBuilder defines how a container is built and pushed to a container registry using the supplied app context.
type ContainerBuilder interface {
	Build(ctx context.Context, app *AppContext, out chan<- *Summary) error
}

// Logs returns the path to the build logs.
//
// Set after Up is called (otherwise "").
func (b *Builder) Logs(appName string) string {
	return filepath.Join(b.LogsDir, appName, b.ID)
}

// Context contains information about the application
type Context struct {
	manifest.Manifest
	AppDir         string
	CustomTags     []string
	DockerContexts []*DockerContext
}

// DockerContext contains information about how docker should build the container
type DockerContext struct {
	Name         string
	Images       []string
	Dockerfile   string
	BuildContext io.ReadCloser
}

// AppContext contains state information carried across the various duffle stage boundaries.
type AppContext struct {
	Bldr           *Builder
	Ctx            *Context
	DockerContexts []*DockerContext
	Log            io.WriteCloser
	ID             string
}

// New creates a new Builder.
func New() *Builder {
	return &Builder{
		ID: getulid(),
	}
}

// newAppContext prepares state carried across the various duffle stage boundaries.
func newAppContext(b *Builder, buildCtx *Context) (*AppContext, error) {
	var buildContexts []*DockerContext

	for _, dockerBuildContext := range buildCtx.DockerContexts {
		defer dockerBuildContext.BuildContext.Close()
		// write each build context to a buffer so we can also write to the sha256 hash.
		buf := new(bytes.Buffer)
		h := sha256.New()
		w := io.MultiWriter(buf, h)
		if _, err := io.Copy(w, dockerBuildContext.BuildContext); err != nil {
			return nil, err
		}
		// truncate checksum to the first 40 characters (20 bytes) this is the
		// equivalent of `shasum build.tar.gz | awk '{print $1}'`.
		ctxtID := h.Sum(nil)
		imgtag := fmt.Sprintf("%.20x", ctxtID)
		imageRepository := path.Join(buildCtx.Registry, fmt.Sprintf("%s-%s", buildCtx.Name, dockerBuildContext.Name))
		image := fmt.Sprintf("%s:%s", imageRepository, imgtag)

		dockerBuildContext.Images = []string{image}
		for _, tag := range buildCtx.CustomTags {
			dockerBuildContext.Images = append(dockerBuildContext.Images, fmt.Sprintf("%s:%s", imageRepository, tag))
		}
		dockerBuildContext.BuildContext = ioutil.NopCloser(buf)
		buildContexts = append(buildContexts, dockerBuildContext)
	}

	if err := osutil.EnsureDirectory(filepath.Dir(b.Logs(buildCtx.Name))); err != nil {
		return nil, err
	}

	logf, err := os.OpenFile(b.Logs(buildCtx.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return &AppContext{
		ID:             b.ID,
		Bldr:           b,
		Ctx:            buildCtx,
		DockerContexts: buildContexts,
		Log:            logf,
	}, nil
}

// LoadWithEnv takes the directory of the application and the environment the application
//  will be pushed to and returns a Context object with a merge of environment and app
//  information
func LoadWithEnv(appdir string) (*Context, error) {
	ctx := &Context{AppDir: appdir}
	// read duffle.toml from appdir.
	tomlFilepath := filepath.Join(appdir, duffle.DuffleTomlFilepath)
	mfst, err := manifest.Load(tomlFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %v", tomlFilepath, err)
	}
	ctx.Manifest = *mfst
	// load the build archives
	if err := loadArchive(ctx); err != nil {
		return nil, fmt.Errorf("failed to load build contexts: %v", err)
	}
	return ctx, nil
}

// loadArchive loads the helm chart and build archive.
func loadArchive(ctx *Context) (err error) {
	for _, component := range ctx.Components {
		dCtx, err := archiveSrc(filepath.Join(ctx.AppDir, component), "")
		if err != nil {
			return err
		}
		ctx.DockerContexts = append(ctx.DockerContexts, dCtx)
	}
	return nil
}

func archiveSrc(contextPath, dockerfileName string) (*DockerContext, error) {
	contextDir, relDockerfile, err := build.GetContextFromLocalDir(contextPath, dockerfileName)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare docker context: %s", err)
	}
	// canonicalize dockerfile name to a platform-independent one
	relDockerfile = archive.CanonicalTarNameForPath(relDockerfile)

	f, err := os.Open(filepath.Join(contextDir, DockerignoreFilename))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	defer f.Close()

	var excludes []string
	if err == nil {
		excludes, err = dockerignore.ReadAll(f)
		if err != nil {
			return nil, err
		}
	}

	if err := build.ValidateContextDirectory(contextDir, excludes); err != nil {
		return nil, fmt.Errorf("error checking docker context: '%s'", err)
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed. The daemon will remove them for us, if needed, after it
	// parses the Dockerfile. Ignore errors here, as they will have been
	// caught by validateContextDirectory above.
	var includes = []string{"."}
	keepThem1, _ := fileutils.Matches(DockerignoreFilename, excludes)
	keepThem2, _ := fileutils.Matches(relDockerfile, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, DockerignoreFilename, relDockerfile)
	}

	logrus.Debugf("INCLUDES: %v", includes)
	logrus.Debugf("EXCLUDES: %v", excludes)
	dockerArchive, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	})
	if err != nil {
		return nil, err
	}

	return &DockerContext{Name: filepath.Base(contextDir), BuildContext: dockerArchive, Dockerfile: relDockerfile}, nil
}

// Build handles incoming duffle build requests and returns a stream of summaries or error.
func (b *Builder) Build(ctx context.Context, bctx *Context) <-chan *Summary {
	ch := make(chan *Summary, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var (
			app *AppContext
			err error
		)
		defer wg.Done()
		if app, err = newAppContext(b, bctx); err != nil {
			log.Printf("error creating app context: %v\n", err)
			return
		}
		log.SetOutput(app.Log)
		if err := b.ContainerBuilder.Build(ctx, app, ch); err != nil {
			log.Printf("error while building: %v\n", err)
			return
		}
	}()
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
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
