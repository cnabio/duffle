package docker

import (
	"fmt"
	"sync"
	"time"

	"github.com/deis/duffle/pkg/builder"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/net/context"
)

// Builder contains information about the build environment
type Builder struct {
	DockerClient command.Cli
}

// Build builds the docker images.
func (b *Builder) Build(ctx context.Context, app *builder.AppContext, out chan<- *builder.Summary) (err error) {
	const stageDesc = "Building Docker Images"

	defer builder.Complete(app.ID, stageDesc, out, &err)
	summary := builder.Summarize(app.ID, stageDesc, out)

	// notify that particular stage has started.
	summary("started", builder.SummaryOngoing)

	errc := make(chan error)
	go func() {
		defer close(errc)
		var wg sync.WaitGroup
		wg.Add(len(app.DockerContexts))
		for _, dockerContext := range app.DockerContexts {
			go func(buildContext *builder.DockerContext) {
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
			summary("ongoing", builder.SummaryOngoing)
			time.Sleep(time.Second)
		}
	}
	return nil
}

// Push pushes the results of Build to the image repository.
func (b *Builder) Push(ctx context.Context, app *builder.AppContext, out chan<- *builder.Summary) (err error) {
	const stageDesc = "Pushing Docker Images"
	if app.Ctx.Registry == "" {
		return
	}

	summary := builder.Summarize(app.ID, stageDesc, out)
	defer builder.Complete(app.ID, stageDesc, out, &err)

	// notify that particular stage has started.
	summary("started", builder.SummaryStarted)

	errc := make(chan error, 1)
	go func() {
		defer close(errc)
		var wg sync.WaitGroup
		wg.Add(len(app.DockerContexts))
		for _, dockerContext := range app.DockerContexts {
			go func(buildContext *builder.DockerContext) {
				defer wg.Done()
				registryAuth, err := command.RetrieveAuthTokenFromImage(ctx, b.DockerClient, buildContext.Images[0])
				if err != nil {
					errc <- err
					return
				}

				for _, tag := range buildContext.Images {
					wg.Add(1)
					go func(tag string) {
						defer wg.Done()
						resp, err := b.DockerClient.Client().ImagePush(ctx, tag, types.ImagePushOptions{RegistryAuth: registryAuth})
						if err != nil {
							errc <- err
							return
						}

						defer resp.Close()
						outFd, isTerm := term.GetFdInfo(app.Log)
						if err := jsonmessage.DisplayJSONMessagesStream(resp, app.Log, outFd, isTerm, nil); err != nil {
							errc <- err
							return
						}
					}(tag)
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
			summary("ongoing", builder.SummaryStarted)
			time.Sleep(time.Second)
		}
	}
	return nil
}
