package auth

import (
	"fmt"
	"testing"
)

func TestAdd(t *testing.T) {
	idx := RepositoryIndex{}
	idx.Add("foo", RepositoryCredentials{})
	if _, ok := idx["foo"]; !ok {
		t.Error("expected idx['foo'] to exist")
	}
	fmt.Println(idx)

	idx.Add("https://example.com", RepositoryCredentials{})
	if _, ok := idx["example.com"]; !ok {
		t.Error("expected idx['example.com'] to exist")
	}

	idx.Add("https://bar.com/foo", RepositoryCredentials{})
	if _, ok := idx["bar.com"]; !ok {
		t.Error("expected idx['bar.com'] to exist")
	}
}

func TestHas(t *testing.T) {
	idx := RepositoryIndex{
		"foo": RepositoryCredentials{},
	}

	if !idx.Has("foo") {
		t.Error("expected 'foo' to exist")
	}
}

func TestGet(t *testing.T) {
	expected := RepositoryCredentials{Token: "hi!"}
	idx := RepositoryIndex{
		"foo": expected,
	}

	creds, err := idx.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if creds != expected {
		t.Errorf("got '%s', expected '%s'", creds, expected)
	}
}

func TestRemove(t *testing.T) {
	idx := RepositoryIndex{
		"foo": RepositoryCredentials{},
	}

	idx.Remove("foo")

	if idx.Has("foo") {
		t.Error("expected 'foo' to not exist")
	}
}
