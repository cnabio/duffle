package digest

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"

	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/client"
	"github.com/docker/docker/registry"

	digest "github.com/opencontainers/go-digest"
)

type dockerDigestValidator struct {
}

// Validate takes a digest and an image reference and fetches the contents to perform
// hash validation on the layers of the image. Inspired by:
// https://gist.github.com/cpuguy83/541dc445fad44193068a1f8f365a9c0e#file-pull-go-L61
func (d *dockerDigestValidator) Validate(ctx context.Context, digest string, image string) error {

	cli := command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false)
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return err
	}
	result, _, err := cli.Client().ImageInspectWithRaw(ctx, image)
	if client.IsErrNotFound(err) {
		ref, err := reference.ParseNormalizedNamed(image)
		if err != nil {
			return err
		}
		// Resolve the Repository name from fqn to RepositoryInfo
		repoInfo, err := registry.ParseRepositoryInfo(ref)
		if err != nil {
			return err
		}

		authConfig := command.ResolveAuthConfig(ctx, cli, repoInfo.Index)
		encodedAuth, err := command.EncodeAuthToBase64(authConfig)
		if err != nil {
			return err
		}
		dist, err := cli.Client().DistributionInspect(ctx, image, encodedAuth)
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
		resolver := docker.NewResolver(docker.ResolverOptions{
			Credentials: func(host string) (string, string, error) {
				return authConfig.Username, authConfig.Password, nil
			},
		})
		name, desc, err := resolver.Resolve(ctx, ref.String())
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
	repoInfo, err := getRepoInfo(image)
	if err != nil {
		return fmt.Errorf("unable to obtain repository info for image: %s", err)
	}
	for _, candidate := range result.RepoDigests {
		candidateRepoInfo, err := getRepoInfo(candidate)
		if err != nil {
			return err
		}
		if candidateRepoInfo.Index.Name == repoInfo.Index.Name {
			fmt.Printf("matched the repo")
			d, err := getDigestInfo(candidate)
			if err != nil {
				return fmt.Errorf("unable to obatin digest for suspected digest: %s", err)
			}
			if string(d) == digest {
				return nil
			}
		}
	}
	return fmt.Errorf("digest doesn't match image")
}

func getRepoInfo(imageName string) (*registry.RepositoryInfo, error) {
	ref, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return nil, err
	}
	// Resolve the Repository name from fqn to RepositoryInfo
	return registry.ParseRepositoryInfo(ref)
}

func getDigestInfo(ref string) (digest.Digest, error) {
	r, err := reference.ParseAnyReference(ref)
	if err != nil {
		return "", err
	}
	digested, ok := r.(reference.Digested)
	if !ok {
		return "", fmt.Errorf("unable to obtain digest")
	}
	return digested.Digest(), nil
}
