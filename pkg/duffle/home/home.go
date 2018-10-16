// Package home provides tools for working with the files in a Duffle user's home directory.
package home

import (
	"os"
	"path/filepath"
	"runtime"
)

// HomeEnvVar is the env var name that points to Duffle home.
const HomeEnvVar = "DUFFLE_HOME"

// PluginEnvVar defines the plugin environment variable name.
const PluginEnvVar = `DUFFLE_PLUGIN`

// Home describes the location of a CLI configuration.
//
// This helper builds paths relative to a Duffle Home directory.
type Home string

// Path returns Home with elements appended.
func (h Home) Path(elem ...string) string {
	p := []string{h.String()}
	p = append(p, elem...)
	return filepath.Join(p...)
}

// Cache returns the path to the Duffle cache.
func (h Home) Cache() string {
	return h.Path("cache")
}

// DefaultRepository returns the name of the default repository.
func (h Home) DefaultRepository() string {
	return "hub.cnlabs.io"
}

// Logs returns the path to the Duffle logs.
func (h Home) Logs() string {
	return h.Path("logs")
}

// Claims is where claims are stored when the filesystem driver is used.
func (h Home) Claims() string {
	return h.Path("claims")
}

// Credentials are where credentialsets are stored.
func (h Home) Credentials() string {
	return h.Path("credentials")
}

// SecretKeyRing returns the path to the keyring containing private keys.
func (h Home) SecretKeyRing() string {
	return h.Path("secret.ring")
}

// PublicKeyRing returns the path to the keyring containing public keys.
func (h Home) PublicKeyRing() string {
	return h.Path("public.ring")
}

// Plugins returns the path to the Duffle plugins.
func (h Home) Plugins() string {
	plugdirs := os.Getenv(PluginEnvVar)

	if plugdirs == "" {
		plugdirs = h.Path("plugins")
	}

	return plugdirs
}

// String returns Home as a string.
//
// Implements fmt.Stringer.
func (h Home) String() string {
	return string(h)
}

// DefaultHome gives the default value for $(duffle home)
func DefaultHome() string {
	if home := os.Getenv(HomeEnvVar); home != "" {
		return home
	}

	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, ".duffle")
}
