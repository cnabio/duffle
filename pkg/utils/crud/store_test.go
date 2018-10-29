package crud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The main point of these tests is to catch any case where the interface
// changes. But we also provide a mock for testing.
var _ Store = &MockStore{}

func TestMockStore(t *testing.T) {
	s := NewMockStore()
	is := assert.New(t)
	is.NoError(s.Store("test", []byte("data")))
	list, err := s.List()
	is.NoError(err)
	is.Len(list, 1)
	data, err := s.Read("test")
	is.NoError(err)
	is.Equal(data, []byte("data"))

}

type MockStore struct {
	data map[string][]byte
}

func NewMockStore() *MockStore {
	return &MockStore{data: map[string][]byte{}}
}

func (s *MockStore) List() ([]string, error) {
	buf := make([]string, len(s.data))
	i := 0
	for k := range s.data {
		buf[i] = k
		i++
	}
	return buf, nil
}
func (s *MockStore) Store(name string, data []byte) error { s.data[name] = data; return nil }
func (s *MockStore) Read(name string) ([]byte, error)     { return s.data[name], nil }
func (s *MockStore) Delete(name string) error             { delete(s.data, name); return nil }
