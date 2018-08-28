package builder

import (
	"fmt"
	"sync"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"
)

// DockerBuilder contains information about the build environment
type DockerBuilder struct {
	DockerClient command.Cli
}

// Build builds the docker images.
func (b DockerBuilder) Build(ctx context.Context, app *AppContext, out chan<- *Summary) (err error) {
	const stageDesc = "Building Docker Images"

	defer Complete(app.ID, stageDesc, out, &err)
	summary := Summarize(app.ID, stageDesc, out)

	// notify that particular stage has started.
	summary("started", SummaryOngoing)

	errc := make(chan error)
	go func() {
		defer close(errc)
		var wg sync.WaitGroup
		wg.Add(len(app.BundleContexts))
		for _, dockerContext := range app.BundleContexts {
			go func(buildContext *BundleContext) {
				defer func() {
					buildContext.BuildContext.Close()
					wg.Done()
				}()
				buildopts := types.ImageBuildOptions{
					Tags:       buildContext.Images,
					Dockerfile: buildContext.Dockerfile,
				}

				resp, err := b.DockerClient.Client().ImageBuild(ctx, buildContext.BuildContext, buildopts)
				if err != nil {
					errc <- err
					return
				}
				defer resp.Body.Close()
				outFd, isTerm := term.GetFdInfo(buildContext.BuildContext)
				if err := jsonmessage.DisplayJSONMessagesStream(resp.Body, app.Log, outFd, isTerm, nil); err != nil {
					errc <- err
					return
				}
				if _, _, err = b.DockerClient.Client().ImageInspectWithRaw(ctx, buildContext.Images[0]); err != nil {
					if dockerclient.IsErrNotFound(err) {
						errc <- fmt.Errorf("Could not locate image for %s: %v", app.Ctx.Name, err)
						return
					}
					errc <- fmt.Errorf("ImageInspectWithRaw error: %v", err)
					return
				}
			}(dockerContext)
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
