package imagestoremocks

import (
	"net/http"

	"github.com/pivotal/image-relocation/pkg/image"
)

type MockStore struct {
	AddStub  func(im string) (string, error)
	PushStub func(image.Digest, image.Name, image.Name) error
}

func (i *MockStore) Add(im string) (string, error) {
	return i.AddStub(im)
}

func (i *MockStore) Push(dig image.Digest, src image.Name, dst image.Name) error {
	return i.PushStub(dig, src, dst)
}

type MockRoundTripper struct{}

func (r *MockRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, nil
}

type MockWriter struct{}

func (w *MockWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}
