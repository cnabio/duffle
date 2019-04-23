package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"

	"github.com/deislabs/duffle/pkg/duffle/home"
)

const (
	testRepositoryPrefix = "example.com/user"

	originalInvocationImageName  = "technosophos/helloworld:0.1.0"
	relocatedInvocationImageName = "example.com/user/technosophos/helloworld/relocated:0.1.0"
	invocationImageDigest        = "sha256:86959ecb500308ae523922eab84f2f94082f20b4e7bda84ce9be219f3f1b4e65"

	originalImageNameA  = "deislabs/duffle@sha256:4d41eeb38fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540"
	relocatedImageNameA = "example.com/user/deislabs/duffle/relocated@sha256:4d41eeb38fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540"
	imageDigestA        = "sha256:4d41eeb38fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540"

	originalImageNameB    = "deislabs/duffle:0.1.0-ralpha.5-englishrose"
	relocatedImageNameB   = "example.com/user/deislabs/duffle/relocated:0.1.0-ralpha.5-englishrose"
	originalImageDigestB  = "sha256:14d6134d892aeccb7e142557fe746ccd0a8f736a747c195ef04c9f3f0f0bbd49"
	relocatedImageDigestB = "sha256:deadbeef892aeccb7e142557fe746ccd0a8f736a747c195ef04c9f3f0f0bbd49"
)

func TestRelocateFileToFilePreservingDigests(t *testing.T) {
	relocateFileToFile(t, true, nil)
}

func TestRelocateFileToFileMutatingDigests(t *testing.T) {
	relocateFileToFile(t, false, fmt.Errorf("digest of image %s not preserved: old digest %s; new digest %s", originalImageNameB, originalImageDigestB, relocatedImageDigestB))
}

func relocateFileToFile(t *testing.T, preserveDigest bool, expectedErr error) {

	duffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(duffleHome)

	work, err := ioutil.TempDir("", "relocatetest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(work)

	outputBundle := filepath.Join(work, "relocated.json")

	cmd := &relocateCmd{
		inputBundle:  "testdata/relocate/bundle.json",
		outputBundle: outputBundle,

		repoPrefix:         testRepositoryPrefix,
		inputBundleIsFile:  true,
		outputBundleIsFile: true,

		home: home.Home(duffleHome),
		out:  ioutil.Discard,

		mapping: func(repoPrefix string, originalImage image.Name) image.Name {
			if repoPrefix != testRepositoryPrefix {
				t.Fatalf("Unexpected repository prefix %s", repoPrefix)
			}
			return testMapping(originalImage, t)
		},
		registryClient: &mockRegClient{
			copyStub: func(source image.Name, target image.Name) (image.Digest, error) {
				oiin, err := image.NewName(originalInvocationImageName)
				if err != nil {
					t.Fatal(err)
				}
				oinA, err := image.NewName(originalImageNameA)
				if err != nil {
					t.Fatal(err)
				}
				oinB, err := image.NewName(originalImageNameB)
				if err != nil {
					t.Fatal(err)
				}
				switch source {
				case oiin:
					if target.String() == relocatedInvocationImageName {
						return image.NewDigest(invocationImageDigest)
					}
				case oinA:
					if target.String() == relocatedImageNameA {
						return image.NewDigest(imageDigestA)
					}
				case oinB:
					if target.String() == relocatedImageNameB {
						if preserveDigest {
							return image.NewDigest(originalImageDigestB)
						}
						// check behaviour if digest is modified, even though this is not normally expected
						return image.NewDigest(relocatedImageDigestB)
					}
				default:
					t.Fatalf("unexpected source %v", source)
				}
				t.Fatalf("unexpected mapping from %v to %v", source, target)
				return image.EmptyDigest, nil // unreachable
			},
		},
	}

	if err := cmd.run(); (err != nil && expectedErr != nil && err.Error() != expectedErr.Error()) ||
		((err == nil || expectedErr == nil) && err != expectedErr) {
		t.Fatalf("unexpected error %v (expected %v)", err, expectedErr)
	}

	if expectedErr != nil {
		return
	}

	// check output bundle
	bundleFile, err := resolveBundleFilePath(outputBundle, "", true)
	if err != nil {
		t.Fatal(err)
	}

	bun, err := loadBundle(bundleFile)
	if err != nil {
		t.Fatal(err)
	}

	if err = bun.Validate(); err != nil {
		t.Fatal(err)
	}

	ii := bun.InvocationImages[0]
	actualInvocationImageName := ii.Image
	if actualInvocationImageName != relocatedInvocationImageName {
		t.Fatalf("output bundle has invocation image with unexpected name: %q (expected %q)",
			actualInvocationImageName, relocatedInvocationImageName)
	}

	actualOriginalInvocationImageName := ii.OriginalImage
	if actualOriginalInvocationImageName != originalInvocationImageName {
		t.Fatalf("output bundle has invocation image with unexpected original name: %q (expected %q)",
			actualOriginalInvocationImageName, originalInvocationImageName)
	}

	assertImage := func(i string, expectedOriginalImageName string, expectedImageName string) {
		img := bun.Images[i]

		actualImageName := img.Image
		if actualImageName != expectedImageName {
			t.Fatalf("output bundle has image %s with unexpected name: %q (expected %q)", i, actualImageName,
				expectedImageName)
		}

		actualOriginalImageName := img.OriginalImage
		if actualOriginalImageName != expectedOriginalImageName {
			t.Fatalf("output bundle has image %s with unexpected original name: %q (expected %q)", i,
				actualOriginalImageName, expectedOriginalImageName)
		}
	}

	assertImage("a", originalImageNameA, relocatedImageNameA)
	assertImage("b", originalImageNameB, relocatedImageNameB)
	assertImage("c", "", "c")
}

// naïve test mapping, preserving any tag and/or digest
// Note: unlike the real mapping, this produces names with more than two slash-separated components.
func testMapping(originalImage image.Name, t *testing.T) image.Name {
	rn, err := image.NewName(path.Join(testRepositoryPrefix, originalImage.Path(), "relocated"))
	if err != nil {
		t.Fatal(err)
	}
	if tag := originalImage.Tag(); tag != "" {
		rn, err = rn.WithTag(tag)
		if err != nil {
			t.Fatal(err)
		}
	}
	if dig := originalImage.Digest(); dig != image.EmptyDigest {
		rn, err = rn.WithDigest(dig)
		if err != nil {
			t.Fatal(err)
		}
	}
	return rn
}

type mockRegClient struct {
	copyStub       func(source image.Name, target image.Name) (image.Digest, error)
	digestStub     func(n image.Name) (image.Digest, error)
	newLayoutStub  func(path string) (registry.Layout, error)
	readLayoutStub func(path string) (registry.Layout, error)
}

func (r *mockRegClient) Digest(n image.Name) (image.Digest, error) { return r.digestStub(n) }
func (r *mockRegClient) Copy(src image.Name, tgt image.Name) (image.Digest, error) {
	return r.copyStub(src, tgt)
}
func (r *mockRegClient) NewLayout(path string) (registry.Layout, error) { return r.newLayoutStub(path) }
func (r *mockRegClient) ReadLayout(path string) (registry.Layout, error) {
	return r.readLayoutStub(path)
}
