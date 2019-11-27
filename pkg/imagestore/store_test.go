package imagestore

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deislabs/duffle/pkg/imagestore/imagestoremocks"
)

func TestCreateParams(t *testing.T) {
	var (
		myLogWriter = &imagestoremocks.MockWriter{}
		myTransport = &imagestoremocks.MockRoundTripper{}
	)

	tests := map[string]struct {
		in  []Option
		out Parameters
	}{
		"defaults": {
			in:  nil,
			out: Parameters{"", ioutil.Discard, http.DefaultTransport},
		},
		"custom log writer": {
			in: []Option{
				WithLogs(myLogWriter),
			},
			out: Parameters{
				"", myLogWriter, http.DefaultTransport,
			},
		},
		"custom transport": {
			in: []Option{
				WithTransport(myTransport),
			},
			out: Parameters{
				"", ioutil.Discard, myTransport,
			},
		},
		"multiple options": {
			in: []Option{
				WithTransport(myTransport),
				WithLogs(myLogWriter),
			},
			out: Parameters{
				"", myLogWriter, myTransport,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.out, CreateParams(tc.in...))
		})
	}
}
