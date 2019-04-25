package relocator

import (
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
)

// FIXME: write this test

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

