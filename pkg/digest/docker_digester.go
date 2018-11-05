package digest

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/client"
)

type dockerDigestValidator struct {
}

// Validate takes a digest and an image reference and fetches the contents to perform
// hash validation on the layers of the image. Inspired by:
// https://gist.github.com/cpuguy83/541dc445fad44193068a1f8f365a9c0e#file-pull-go-L61
func (d *dockerDigestValidator) Validate(ctx context.Context, digest string, image string) error {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("cannot create Docker client: %v", err)
	}
	cli.NegotiateAPIVersion(ctx)
	if err != nil {
		return fmt.Errorf("cannot update Docker client: %v", err)
	}
	dist, err := cli.DistributionInspect(ctx, image, "")
	if err != nil {
		return fmt.Errorf("unable to inspect image: %s", err)
	}
	if string(dist.Descriptor.Digest) != digest {
		return fmt.Errorf(
			"digest from registry %s does not match expected: %s",
			dist.Descriptor.Digest,
			digest,
		)
	}
	dir, err := ioutil.TempDir("", "testPull")
	if err != nil {
		return fmt.Errorf("unable to make validation temp directory: %s", err)
	}
	defer os.RemoveAll(dir)
	cs, err := local.NewStore(dir)
	if err != nil {
		return fmt.Errorf("unable to make content store: %s", err)
	}
	// In the case that the image from the bundle isn't normalized (i.e. doesn't include a full domain)
	// use the docker Reference package to normalize it
	normalized, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return fmt.Errorf("unknown image format: %s", err)
	}
	resolver := docker.NewResolver(docker.ResolverOptions{})
	name, desc, err := resolver.Resolve(ctx, normalized.String())
	if err != nil {
		return fmt.Errorf("unable to resolve image: %s", err)
	}
	fetcher, err := resolver.Fetcher(ctx, name)
	if err != nil {
		return fmt.Errorf("unable to create fetcher: %s", err)
	}
	r, err := fetcher.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("unable to fetch image descriptor: %s", err)
	}
	defer r.Close()
	// Handler which reads a descriptor and fetches the referenced data (e.g. image layers) from the remote
	h := remotes.FetchHandler(cs, fetcher)
	// This traverses the OCI descriptor to fetch the image and store it into the local store initialized above.
	// All content hashes are verified in this step
	if err := images.Dispatch(ctx, h, desc); err != nil {
		return fmt.Errorf("error verifying image: %s", err)
	}
	return nil
}
