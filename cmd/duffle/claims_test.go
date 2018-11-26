package main

import (
	"errors"
	"testing"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/claim"

	"github.com/stretchr/testify/assert"
)

func TestMockClaimStore(t *testing.T) {
	is := assert.New(t)
	store := mockClaimStore()
	c := claim.Claim{
		Name: "testclaim",
		Bundle: &bundle.Bundle{
			Name: "testbundle",
		},
	}
	store.Store(c)
	list, err := store.List()
	is.NoError(err)
	is.Len(list, 1)
	got, err := store.Read("testclaim")
	is.NoError(err)
	is.Equal("testbundle", got.Bundle.Name)
}

func mockClaimStore() claim.Store {
	return claim.NewClaimStore(mockClaimBackend{})
}

type mockClaimBackend map[string][]byte

func (m mockClaimBackend) List() ([]string, error) {
	list := []string{}
	for key := range m {
		list = append(list, key)
	}
	return list, nil
}

func (m mockClaimBackend) Store(name string, data []byte) error {
	m[name] = data
	return nil
}

func (m mockClaimBackend) Read(name string) ([]byte, error) {
	data, ok := m[name]
	if !ok {
		return data, errors.New("not found")
	}
	return data, nil
}

func (m mockClaimBackend) Delete(name string) error {
	delete(m, name)
	return nil
}
