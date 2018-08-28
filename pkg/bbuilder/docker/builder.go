package docker

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
	"strings"
	"sync"

	"github.com/deis/duffle/pkg/osutil"

	"github.com/deis/duffle/pkg/bbuilder"
	"github.com/deis/duffle/pkg/builder"
	"github.com/deis/duffle/pkg/duffle"
	"github.com/deis/duffle/pkg/duffle/manifest"
	"github.com/docker/cli/cli/command"
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

// Component contains all information to build a container image
type Component struct {
	Name         string
	Image        string
	Dockerfile   string
	BuildContext io.ReadCloser
}

var _ bbuilder.Component = (*Component)(nil)
var _ bbuilder.BundleBuilder = (*Builder)(nil)

// URI returns the image in the format <registry>/<image>
func (dc Component) URI() string {
	return dc.Image
}

// Digest returns the name of a Docker component, which will give the image name
//
// TODO - return the actual digest
func (dc Component) Digest() string {
	return strings.Split(dc.Image, ":")[1]
}

// Builder contains information about the Docker build environment
type Builder struct {
	DockerClient command.Cli
}

// PrepareBuild prepares state carried across the various duffle stage boundaries.
func (d Builder) PrepareBuild(bldr *bbuilder.Builder, appDir string) (*bbuilder.AppContext, error) {

	ctx, err := loadContext(appDir)
	if err != nil {
		return nil, fmt.Errorf("cannot load app context: %v", err)
	}

	for _, c := range ctx.Components {
		dc, ok := c.(*Component)
		if !ok {
			return nil, fmt.Errorf("cannot convert component to Docker component in prepare")
		}

		defer dc.BuildContext.Close()

		// write each build context to a buffer so we can also write to the sha256 hash.
		buf := new(bytes.Buffer)
		h := sha256.New()
		w := io.MultiWriter(buf, h)
		if _, err := io.Copy(w, dc.BuildContext); err != nil {
			return nil, err
		}

		// truncate checksum to the first 40 characters (20 bytes) this is the
		// equivalent of `shasum build.tar.gz | awk '{print $1}'`.
		ctxtID := h.Sum(nil)
		imgtag := fmt.Sprintf("%.20x", ctxtID)
		imageRepository := path.Join(ctx.Manifest.Registry, fmt.Sprintf("%s-%s", ctx.Manifest.Name, dc.Name))
		dc.Image = fmt.Sprintf("%s:%s", imageRepository, imgtag)

		dc.BuildContext = ioutil.NopCloser(buf)
		ctx.Components = append(ctx.Components, dc)
	}

	if err := osutil.EnsureDirectory(filepath.Dir(bldr.Logs(ctx.Manifest.Name))); err != nil {
		return nil, err
	}

	logf, err := os.OpenFile(bldr.Logs(ctx.Manifest.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return &bbuilder.AppContext{
		ID:   bldr.ID,
		Bldr: bldr,
		Ctx:  ctx,
		Log:  logf,
	}, nil
}

// Build builds the docker images.
func (d Builder) Build(ctx context.Context, app *bbuilder.AppContext, out chan<- *builder.Summary) (err error) {
	ch := make(chan *builder.Summary, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(app *bbuilder.AppContext) {
		defer wg.Done()
		log.SetOutput(app.Log)
		if err := d.BuildComponents(ctx, app, ch); err != nil {
			log.Printf("error while building: %v\n", err)
			return
		}
	}(app)
	go func() {
		wg.Wait()
		close(ch)
	}()

	return nil
}

func loadContext(appDir string) (*bbuilder.Context, error) {
	ctx := &bbuilder.Context{AppDir: appDir}

	tomlFilePath := filepath.Join(appDir, duffle.DuffleTomlFilepath)
	mfst, err := manifest.Load(tomlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %v", tomlFilePath, err)
	}
	ctx.Manifest = *mfst

	if err := loadArchive(ctx); err != nil {
		return nil, fmt.Errorf("failed to load build contexts: %v", err)
	}

	return ctx, nil
}

// loadArchive loads the helm chart and build archive.
func loadArchive(ctx *bbuilder.Context) (err error) {
	for _, component := range ctx.Manifest.Components {
		dc, err := archiveSrc(filepath.Join(ctx.AppDir, component), "")
		if err != nil {
			return err
		}
		ctx.Components = append(ctx.Components, dc)
	}
	return nil
}

func archiveSrc(contextPath, dockerfileName string) (*Component, error) {
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

	return &Component{Name: filepath.Base(contextDir), BuildContext: dockerArchive, Dockerfile: relDockerfile}, nil
}
