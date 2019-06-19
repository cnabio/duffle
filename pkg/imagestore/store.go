package imagestore

import "github.com/pivotal/image-relocation/pkg/image"

// Store is an abstract image store.
type Store interface {
	// Add copies the image with the given name to the image store.
	Add(img string) (contentDigest string, err error)

	// Push copies the image with the given digest from an image with the given name in the image store to a repository
	// with the given name.
	Push(dig image.Digest, src image.Name, dst image.Name) error
}
