package loader

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testFooJSON = filepath.Join("..", "..", "tests", "testdata", "bundles", "foo.json")

func TestLoader(t *testing.T) {
	is := assert.New(t)

	l := NewLoader()
	bundle, err := l.Load(testFooJSON)
	if err != nil {
		t.Fatalf("cannot load bundle: %v", err)
	}

	is.Equal("foo", bundle.Name)
	is.Equal("1.0", bundle.Version)
}
func TestLoader_Remote(t *testing.T) {

	data, err := ioutil.ReadFile(testFooJSON)
	if err != nil {
		t.Fatalf("cannot read bundle file: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, bytes.NewBuffer(data))
	}))
	defer ts.Close()

	l := NewLoader()
	bundle, err := l.Load(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	is := assert.New(t)

	is.Equal("foo", bundle.Name)
	is.Equal("1.0", bundle.Version)
}
