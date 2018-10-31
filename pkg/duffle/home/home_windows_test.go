// +build windows

package home

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuffleHome(t *testing.T) {
	is := assert.New(t)
	ph := Home("r:\\")
	is.Equal(ph.String(), "r:\\")
	is.Equal(ph.Plugins(), "r:\\plugins")
	is.Equal(ph.Bundles(), "r:\\bundles")
	is.Equal(ph.Claims(), "r:\\claims")
	is.Equal(ph.Credentials(), "r:\\credentials")
	is.Equal(ph.Logs(), "r:\\logs")
	is.Equal(ph.Repositories(), "r:\\repositories.json")
	is.Equal(ph.SecretKeyRing(), "r:\\secret.ring")
	is.Equal(ph.PublicKeyRing(), "r:\\public.ring")
}
