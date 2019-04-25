package relocator_test

import (
	"fmt"
	"testing"

	"github.com/pivotal/image-relocation/pkg/image"

	"github.com/deislabs/duffle/pkg/imagestore/imagestoremocks"
	"github.com/deislabs/duffle/pkg/relocator"

	"github.com/deislabs/cnab-go/bundle"

	"github.com/deislabs/duffle/pkg/imagestore"
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
		name     string
		fields   fields
		wantErr  bool
		checkBun func(*bundle.Bundle)
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
			func(bun *bundle.Bundle) {
				im := bun.InvocationImages[0]
				if im.Image != expectedMappedRef {
					t.Errorf("expected mapped image %s, got %s", expectedMappedRef, im.Image)
				}
				if im.OriginalImage != expectedOriginalShortRef {
					t.Errorf("expected original image %s, got %s", expectedOriginalShortRef, im.OriginalImage)
				}
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
			func(bun *bundle.Bundle) {
				im := bun.Images["i1"]
				if im.Image != expectedMappedRef {
					t.Errorf("expected mapped image %s, got %s", expectedMappedRef, im.Image)
				}
				if im.OriginalImage != expectedOriginalShortRef {
					t.Errorf("expected original image %s, got %s", expectedOriginalShortRef, im.OriginalImage)
				}
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
			func(bun *bundle.Bundle) {
				im := bun.Images["i1"]
				if im.Image != expectedMappedRef {
					t.Errorf("expected mapped image %s, got %s", expectedMappedRef, im.Image)
				}
				if im.OriginalImage != expectedOriginalShortDigestedRef {
					t.Errorf("expected original image %s, got %s", expectedOriginalShortDigestedRef, im.OriginalImage)
				}
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
			func(bun *bundle.Bundle) {
				im := bun.Images["i1"]
				if im.Image != expectedMappedRef {
					t.Errorf("expected mapped image %s, got %s", expectedMappedRef, im.Image)
				}
				if im.OriginalImage != expectedOriginalShortRef {
					t.Errorf("expected original image %s, got %s", expectedOriginalShortRef, im.OriginalImage)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := relocator.NewRelocator(tt.fields.bun, tt.fields.mapping, tt.fields.imageStore)
			if err != nil {
				t.Fatalf("NewRelocate failed: %v", err)
			}
			if err := r.Relocate(); (err != nil) != tt.wantErr {
				t.Errorf("Relocator.Relocate() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.checkBun(tt.fields.bun)
		})
	}
}
