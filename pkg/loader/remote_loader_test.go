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

func TestRmoteLoader(t *testing.T) {

	data, err := ioutil.ReadFile(filepath.Join("..", "..", "tests", "testdata", "bundles", "foo.json"))
	if err != nil {
		t.Fatalf("cannot read bundle file: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, bytes.NewBuffer(data))
	}))
	defer ts.Close()

	l := RemoteLoader{source: ts.URL}

	bundle, err := l.Load()
	if err != nil {
		t.Fatal(err)
	}
	is := assert.New(t)

	is.Equal("foo", bundle.Name)
	is.Equal("1.0", bundle.Version)
}
