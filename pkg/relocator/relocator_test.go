package relocator_test

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/pivotal/image-relocation/pkg/image"

	"github.com/cnabio/duffle/pkg/imagestore/imagestoremocks"
	"github.com/cnabio/duffle/pkg/relocator"

	"github.com/deislabs/cnab-go/bundle"

	"github.com/cnabio/duffle/pkg/imagestore"
)

func TestRelocator_Relocate(t *testing.T) {
	const (
		expectedOriginalShortRef         = "ubuntu"
		expectedOriginalRef              = "docker.io/library/ubuntu"
		expectedMappedRef                = "docker.io/library/ubuntu-mapped"
		expectedDigest                   = "sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
		expectedOriginalShortDigestedRef = expectedOriginalShortRef + "@" + expectedDigest
		expectedOriginalDigestedRef      = expectedOriginalRef + "@" + expectedDigest
	)

	type fields struct {
		bun        *bundle.Bundle
		mapping    relocator.Mapping
		imageStore imagestore.Store
	}
	tests := []struct {
		name           string
		fields         fields
		wantErr        bool
		expectedRelMap map[string]string
	}{
		{
			"invocation image",
			fields{
				&bundle.Bundle{
					InvocationImages: []bundle.InvocationImage{
						{
							BaseImage: bundle.BaseImage{
								Image: expectedOriginalShortRef,
							},
						},
					},
				},
				func(ref image.Name) image.Name {
					mappedRef, err := image.NewName(fmt.Sprintf("%s/%s-mapped", ref.Host(), ref.Path()))
					if err != nil {
						t.Fatalf("Unexpected error: %v", err)
					}
					return mappedRef
				},
				&imagestoremocks.MockStore{
					PushStub: func(dig image.Digest, src image.Name, dst image.Name) error {
						if dig != image.EmptyDigest {
							t.Errorf("expected digest %s, got %s", image.EmptyDigest, dig)
						}
						if src.String() != expectedOriginalRef {
							t.Errorf("expected source image %s, got %s", expectedOriginalRef, src.String())
						}
						if dst.String() != expectedMappedRef {
							t.Errorf("expected destination image %s, got %s", expectedMappedRef, dst.String())
						}
						return nil
					},
				},
			},
			false,
			map[string]string{
				expectedOriginalShortRef: expectedMappedRef,
			},
		},
		{
			"image",
			fields{
				&bundle.Bundle{
					Images: map[string]bundle.Image{
						"i1": {
							BaseImage: bundle.BaseImage{
								Image:     expectedOriginalShortRef,
								ImageType: "docker",
							},
						},
					},
				},
				func(ref image.Name) image.Name {
					mappedRef, err := image.NewName(fmt.Sprintf("%s/%s-mapped", ref.Host(), ref.Path()))
					if err != nil {
						t.Fatalf("Unexpected error: %v", err)
					}
					return mappedRef
				},
				&imagestoremocks.MockStore{
					PushStub: func(dig image.Digest, src image.Name, dst image.Name) error {
						if dig != image.EmptyDigest {
							t.Errorf("expected digest %s, got %s", image.EmptyDigest, dig)
						}

						if src.String() != expectedOriginalRef {
							t.Errorf("expected source image %s, got %s", expectedOriginalRef, src.String())
						}

						if dst.String() != expectedMappedRef {
							t.Errorf("expected destination image %s, got %s", expectedMappedRef, dst.String())
						}
						return nil
					},
				},
			},
			false,
			map[string]string{
				expectedOriginalShortRef: expectedMappedRef,
			},
		},
		{
			"digested image",
			fields{
				&bundle.Bundle{
					Images: map[string]bundle.Image{
						"i1": {
							BaseImage: bundle.BaseImage{
								Image: expectedOriginalShortDigestedRef,
							},
						},
					},
				},
				func(ref image.Name) image.Name {
					mappedRef, err := image.NewName(fmt.Sprintf("%s/%s-mapped", ref.Host(), ref.Path()))
					if err != nil {
						t.Fatalf("Unexpected error: %v", err)
					}
					return mappedRef
				},
				&imagestoremocks.MockStore{
					PushStub: func(dig image.Digest, src image.Name, dst image.Name) error {
						if dig.String() != expectedDigest {
							t.Errorf("expected digest %s, got %s", expectedDigest, dig)
						}

						if src.String() != expectedOriginalDigestedRef {
							t.Errorf("expected source image %s, got %s", expectedOriginalDigestedRef, src.String())
						}

						if dst.String() != expectedMappedRef {
							t.Errorf("expected destination image %s, got %s", expectedMappedRef, dst.String())
						}
						return nil
					},
				},
			},
			false,
			map[string]string{
				expectedOriginalShortDigestedRef: expectedMappedRef,
			},
		},
		{
			"image with declared digest",
			fields{
				&bundle.Bundle{
					Images: map[string]bundle.Image{
						"i1": {
							BaseImage: bundle.BaseImage{
								Image:  expectedOriginalShortRef,
								Digest: expectedDigest,
							},
						},
					},
				},
				func(ref image.Name) image.Name {
					mappedRef, err := image.NewName(fmt.Sprintf("%s/%s-mapped", ref.Host(), ref.Path()))
					if err != nil {
						t.Fatalf("Unexpected error: %v", err)
					}
					return mappedRef
				},
				&imagestoremocks.MockStore{
					PushStub: func(dig image.Digest, src image.Name, dst image.Name) error {
						if dig.String() != expectedDigest {
							t.Errorf("expected digest %s, got %s", expectedDigest, dig)
						}

						if src.String() != expectedOriginalRef {
							t.Errorf("expected source image %s, got %s", expectedOriginalRef, src.String())
						}

						if dst.String() != expectedMappedRef {
							t.Errorf("expected destination image %s, got %s", expectedMappedRef, dst.String())
						}
						return nil
					},
				},
			},
			false,
			map[string]string{
				expectedOriginalShortRef: expectedMappedRef,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relMap := make(map[string]string)
			r, err := relocator.NewRelocator(tt.fields.bun, tt.fields.mapping, tt.fields.imageStore, ioutil.Discard)
			if err != nil {
				t.Fatalf("NewRelocate failed: %v", err)
			}
			if err := r.Relocate(relMap); (err != nil) != tt.wantErr {
				t.Errorf("Relocator.Relocate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(relMap, tt.expectedRelMap) {
				t.Errorf("output relocation mapping file has unexpected content: %v (expected %v)",
					relMap, tt.expectedRelMap)
			}
		})
	}
}
