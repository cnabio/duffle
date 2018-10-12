// +build !windows

package home

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuffleHome(t *testing.T) {
	is := assert.New(t)
	ph := Home("/r")
	runtime := fmt.Sprintf("runtime %s", runtime.GOOS)
	is.Equal(ph.String(), "/r", runtime)
	is.Equal(ph.Cache(), "/r/cache", runtime)
	is.Equal(ph.Plugins(), "/r/plugins", runtime)
	is.Equal(ph.Claims(), "/r/claims", runtime)
	is.Equal(ph.Credentials(), "/r/credentials", runtime)
	is.Equal(ph.Logs(), "/r/logs", runtime)
	is.Equal(ph.SecretKeyRing(), "/r/secret.ring", runtime)
	is.Equal(ph.PublicKeyRing(), "/r/public.ring", runtime)
}
