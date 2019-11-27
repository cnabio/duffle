package construction

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deislabs/duffle/pkg/imagestore"
	"github.com/deislabs/duffle/pkg/imagestore/imagestoremocks"
)

func TestNewLocatingConstructor(t *testing.T) {
	var (
		myLogWriter   = &imagestoremocks.MockWriter{}
		myTransport   = &imagestoremocks.MockRoundTripper{}
		myThickBundle = mustCreateThickBundleDir(t)
	)

	defer os.RemoveAll(myThickBundle)

	tests := map[string]struct {
		opts   []imagestore.Option
		expect imagestore.Parameters
	}{
		"thick bundle": {
			opts: []imagestore.Option{
				imagestore.WithArchiveDir(myThickBundle),
			},
			expect: imagestore.Parameters{
				ArchiveDir: myThickBundle,
				Logs:       ioutil.Discard,
				Transport:  http.DefaultTransport,
			},
		},
		"thin bundle - default store options": {
			opts: []imagestore.Option{},
			expect: imagestore.Parameters{
				ArchiveDir: "",
				Logs:       ioutil.Discard,
				Transport:  http.DefaultTransport,
			},
		},
		"thin bundle - custom transport": {
			opts: []imagestore.Option{
				imagestore.WithTransport(myTransport),
			},
			expect: imagestore.Parameters{
				ArchiveDir: "",
				Logs:       ioutil.Discard,
				Transport:  myTransport,
			},
		},
		"thin bundle - custom log writer": {
			opts: []imagestore.Option{
				imagestore.WithLogs(myLogWriter),
			},
			expect: imagestore.Parameters{
				ArchiveDir: "",
				Logs:       myLogWriter,
				Transport:  http.DefaultTransport,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				remoteConstructorCalled    = false
				ocilayoutConstructorCalled = false
			)

			locatingConstructorRemote = func(opts ...imagestore.Option) (imagestore.Store, error) {
				remoteConstructorCalled = true
				assert.Equal(t, tc.expect, imagestore.CreateParams(opts...))
				return nil, nil
			}

			locatingConstructorOciLayout = func(opts ...imagestore.Option) (imagestore.Store, error) {
				ocilayoutConstructorCalled = true
				assert.Equal(t, tc.expect, imagestore.CreateParams(opts...))
				return nil, nil
			}

			NewLocatingConstructor()(tc.opts...)

			assert.Equal(t, tc.expect.ArchiveDir == "", remoteConstructorCalled)
			assert.Equal(t, tc.expect.ArchiveDir != "", ocilayoutConstructorCalled)
		})
	}
}

func mustCreateThickBundleDir(t *testing.T) string {
	name, err := ioutil.TempDir("", "bundle")
	if err != nil {
		t.Fatal(err)
	}

	if err = os.Mkdir(filepath.Join(name, "artifacts"), os.ModePerm); err != nil {
		t.Fatal(err)
	}

	return name
}
