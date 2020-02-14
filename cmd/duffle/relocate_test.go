package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/pathmapping"
	"github.com/pivotal/image-relocation/pkg/transport"

	"github.com/cnabio/duffle/pkg/duffle/home"
	"github.com/cnabio/duffle/pkg/imagestore"
	"github.com/cnabio/duffle/pkg/imagestore/construction"
	"github.com/cnabio/duffle/pkg/imagestore/imagestoremocks"
)

const (
	testRepositoryPrefix = "example.com/user"

	originalInvocationImageName  = "technosophos/helloworld:0.1.0"
	relocatedInvocationImageName = "example.com/user/technosophos-helloworld-6731e0d41b7fd5a24e14c853af93bd81:0.1.0"

	originalImageNameA  = "deislabs/duffle@sha256:4d41eeb38fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540"
	relocatedImageNameA = "example.com/user/deislabs-duffle-50aa5cc4ebb040ac696a9753d1695298@sha256:4d41eeb38fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540"
	imageDigestA        = "sha256:4d41eeb38fb14266b7c0461ef1ef0b2f8c05f41cd544987a259a9d92cdad2540"

	originalImageNameB   = "deislabs/duffle:0.1.0-ralpha.5-englishrose"
	relocatedImageNameB  = "example.com/user/deislabs-duffle-50aa5cc4ebb040ac696a9753d1695298:0.1.0-ralpha.5-englishrose"
	originalImageDigestB = "sha256:14d6134d892aeccb7e142557fe746ccd0a8f736a747c195ef04c9f3f0f0bbd49"
)

func TestRelocate(t *testing.T) {
	tests := map[string]struct {
		bundle                  string
		expectedArchiveDirRegex string
		expectedErr             error
	}{
		"from file": {
			"testdata/relocate/bundle.json",
			"^$",
			nil,
		},
		"from file with unsupported image type": {
			"testdata/relocate/bundle-with-unsupported-image-type.json",
			"^$",
			errors.New("cannot relocate image c with imageType c: only oci and docker image types are currently supported"),
		},
		"from archive": {
			"testdata/relocate/testrelocate-0.1.tgz",
			"^.*testrelocate-0\\.1$",
			nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			homedir := mustCreateTempDir(t, "dufflehome")
			defer os.RemoveAll(homedir)

			workdir := mustCreateTempDir(t, "relocatetest")
			defer os.RemoveAll(workdir)

			relMapPath := filepath.Join(workdir, "relmap.json")

			cmd := initRelocateCmd(tc.bundle, homedir, relMapPath)
			cmd.mapping = pathMappingStub(t)
			cmd.imageStoreConstructor = imageStoreConstructorStub(t, tc.expectedArchiveDirRegex)

			err := cmd.run()

			assertErrorMessagesMatch(t, err, tc.expectedErr)

			if err == nil {
				assertRelocationMap(t, relMapPath)
			}
		})
	}
}

func TestRelocateTransportOptions(t *testing.T) {
	tests := map[string]struct {
		expectedCertPaths     []string
		expectedSkipTLSVerify bool
		expectedErr           error
	}{
		"defaults": {
			nil,
			false,
			nil,
		},
		"one cert": {
			[]string{"a"},
			false,
			nil,
		},
		"multiple certs": {
			[]string{"a", "b", "c"},
			false,
			nil,
		},
		"skip TLS verify": {
			nil,
			true,
			nil,
		},
		"construction error": {
			nil,
			false,
			errors.New("i like turtles"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			homedir := mustCreateTempDir(t, "dufflehome")
			defer os.RemoveAll(homedir)

			workdir := mustCreateTempDir(t, "relocatetest")
			defer os.RemoveAll(workdir)

			expectedTransport := &http.Transport{}

			cmd := initRelocateCmd("testdata/relocate/bundle.json", homedir, filepath.Join(workdir, "relmap.json"))

			cmd.caCertPaths = tc.expectedCertPaths
			cmd.skipTLSVerify = tc.expectedSkipTLSVerify

			imageStoreContructorCalled := false
			cmd.imageStoreConstructor = func(opts ...imagestore.Option) (store imagestore.Store, e error) {
				imageStoreContructorCalled = true

				p := imagestore.Parameters{}
				for _, opt := range opts {
					p = opt(p)
				}

				assert.Same(t, expectedTransport, p.Transport)

				return &imagestoremocks.MockStore{
					PushStub: func(image.Digest, image.Name, image.Name) error {
						return nil
					},
				}, nil
			}

			transportContructorCalled := false
			cmd.transportConstructor = func(certPaths []string, skipTLSVerify bool) (*http.Transport, error) {
				transportContructorCalled = true
				assert.ElementsMatch(t, tc.expectedCertPaths, certPaths)
				assert.Equal(t, tc.expectedSkipTLSVerify, skipTLSVerify)
				return expectedTransport, tc.expectedErr
			}

			assertErrorMessagesMatch(t, cmd.run(), tc.expectedErr)
			assert.True(t, transportContructorCalled)
			assert.Equal(t, tc.expectedErr == nil, imageStoreContructorCalled)
		})
	}
}

func mustCreateTempDir(t *testing.T, prefix string) string {
	dirname, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}

	return dirname
}

func initRelocateCmd(bundle, homedir, relMapPath string) *relocateCmd {
	return &relocateCmd{
		inputBundle:           bundle,
		repoPrefix:            testRepositoryPrefix,
		bundleIsFile:          true,
		relocationMapping:     relMapPath,
		home:                  home.Home(homedir),
		out:                   ioutil.Discard,
		mapping:               pathmapping.FlattenRepoPathPreserveTagDigest,
		transportConstructor:  transport.NewHttpTransport,
		imageStoreConstructor: construction.NewLocatingConstructor(),
	}
}

func pathMappingStub(t *testing.T) pathmapping.PathMapping {
	return func(repoPrefix string, originalImage image.Name) image.Name {
		if repoPrefix != testRepositoryPrefix {
			t.Fatalf("Unexpected repository prefix %s", repoPrefix)
		}
		return mockMapping(t, originalImage)
	}
}

func imageStoreConstructorStub(t *testing.T, expectedArchiveDirRegex string) imagestore.Constructor {
	return func(option ...imagestore.Option) (store imagestore.Store, e error) {
		archiveDir := imagestore.CreateParams(option...).ArchiveDir
		assert.Regexp(t, expectedArchiveDirRegex, archiveDir)
		return mockImageStore(t), nil
	}
}

// naïve test mapping, preserving any tag and/or digest
// Note: unlike the real mapping, this produces names with more than two slash-separated components.
func mockMapping(t *testing.T, originalImage image.Name) image.Name {
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

func mockImageStore(t *testing.T) imagestore.Store {
	return &imagestoremocks.MockStore{
		PushStub: func(dig image.Digest, src image.Name, dst image.Name) error {
			type pair struct {
				first  string
				second string
			}
			digests := map[string]pair{
				"docker.io/" + originalInvocationImageName: {"", relocatedInvocationImageName},
				"docker.io/" + originalImageNameA:          {imageDigestA, relocatedImageNameA},
				"docker.io/" + originalImageNameB:          {originalImageDigestB, relocatedImageNameB},
			}
			exp, ok := digests[src.String()]
			if !ok {
				t.Fatalf("unexpected source image %v", src)
			}
			expectedDig, expectedDst := exp.first, exp.second
			if dig.String() != expectedDig {
				t.Fatalf("digest for source image %v was %s, expected %s", src, dig, expectedDig)
			}
			if dst.String() != expectedDst {
				t.Fatalf("destination image for source image %v was %v, expected %s", src, dst, expectedDst)
			}
			return nil
		},
	}
}

func assertErrorMessagesMatch(t *testing.T, actualErr, expectedErr error) {
	if (actualErr != nil && expectedErr != nil && actualErr.Error() != expectedErr.Error()) ||
		((actualErr == nil || expectedErr == nil) && actualErr != expectedErr) {
		t.Fatalf("unexpected error %v (expected %v)", actualErr, expectedErr)
	}
}

func assertRelocationMap(t *testing.T, relMapPath string) {
	data, err := ioutil.ReadFile(relMapPath)
	if err != nil {
		t.Fatal(err)
	}

	relMap := make(map[string]string)

	err = json.Unmarshal(data, &relMap)
	if err != nil {
		t.Fatal(err)
	}

	expectedRelMap := map[string]string{
		originalInvocationImageName: relocatedInvocationImageName,
		originalImageNameA:          relocatedImageNameA,
		originalImageNameB:          relocatedImageNameB,
	}

	if !reflect.DeepEqual(relMap, expectedRelMap) {
		t.Fatalf(
			"output relocation mapping file has unexpected content: %v (expected %v)",
			relMap, expectedRelMap,
		)
	}
}
