// +build !windows

package home

import (
	"runtime"
	"testing"
)

func TestDuffleHome(t *testing.T) {
	ph := Home("/r")
	isEq := func(t *testing.T, a, b string) {
		if a != b {
			t.Error(runtime.GOOS)
			t.Errorf("Expected %q, got %q", a, b)
		}
	}

	isEq(t, ph.String(), "/r")
	isEq(t, ph.Repositories(), "/r/repositories")
	isEq(t, ph.Plugins(), "/r/plugins")
	isEq(t, ph.Claims(), "/r/claims")
	isEq(t, ph.Credentials(), "/r/credentials")
}
