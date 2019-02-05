package crud

import "errors"

// ErrDoesNotExist represents when an item is not found in the store
var ErrDoesNotExist = errors.New(`does not exist`)

// Store is a simplified interface to a key-blob store supporting CRUD operations.
type Store interface {
	List() ([]string, error)
	Store(name string, data []byte) error
	Read(name string) ([]byte, error)
	Delete(name string) error
}
