// +build windows

package home

import (
	"testing"
)

func TestDuffleHome(t *testing.T) {
	ph := Home("r:\\")
	isEq := func(t *testing.T, a, b string) {
		if a != b {
			t.Errorf("Expected %q, got %q", b, a)
		}
	}

	isEq(t, ph.String(), "r:\\")
	isEq(t, ph.Repositories(), "r:\\repositories")
	isEq(t, ph.Plugins(), "r:\\plugins")
}
