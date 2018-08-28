package docker

import (
	"fmt"
	"sync"
	"time"

	"github.com/deis/duffle/pkg/bbuilder"
	"github.com/deis/duffle/pkg/builder"
	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"
)

// BuildComponents builds the docker images.
func (b Builder) BuildComponents(ctx context.Context, app *bbuilder.AppContext, out chan<- *builder.Summary) (err error) {
	const stageDesc = "Building Docker Images"

	defer builder.Complete(app.ID, stageDesc, out, &err)
	summary := builder.Summarize(app.ID, stageDesc, out)

	// notify that particular stage has started.
	summary("started", builder.SummaryOngoing)

	errc := make(chan error)
	go func() {
		defer close(errc)
		var wg sync.WaitGroup
		wg.Add(len(app.Ctx.Components))
		for _, c := range app.Ctx.Components {

			dc, ok := c.(Component)
			if !ok {
				errc <- fmt.Errorf("cannot convert component to Docker component")
			}

			go func(dc *Component) {
				defer func() {
					dc.BuildContext.Close()
					wg.Done()
				}()
				buildopts := types.ImageBuildOptions{
					Tags:       []string{dc.Image},
					Dockerfile: dc.Dockerfile,
				}

				resp, err := b.DockerClient.Client().ImageBuild(ctx, dc.BuildContext, buildopts)
				if err != nil {
					errc <- err
					return
				}
				defer resp.Body.Close()
				outFd, isTerm := term.GetFdInfo(dc.BuildContext)
				if err := jsonmessage.DisplayJSONMessagesStream(resp.Body, app.Log, outFd, isTerm, nil); err != nil {
					errc <- err
					return
				}
				if _, _, err = b.DockerClient.Client().ImageInspectWithRaw(ctx, dc.Image); err != nil {
					if dockerclient.IsErrNotFound(err) {
						errc <- fmt.Errorf("Could not locate image for %s: %v", app.Ctx.Manifest.Name, err)
						return
					}
					errc <- fmt.Errorf("ImageInspectWithRaw error: %v", err)
					return
				}
			}(&dc)
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
			summary("ongoing", builder.SummaryOngoing)
			time.Sleep(time.Second)
		}
	}
	return nil
}
