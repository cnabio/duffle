package manifest

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	m := New()
	expected := "&{manifest  [] [] map[] map[] map[]}"
	actual := fmt.Sprintf("%v", m)
	if expected != actual {
		t.Errorf("wanted %s, got %s", expected, actual)
	}
}

func TestGenerateName(t *testing.T) {
	name := generateName()
	if name == "" {
		t.Error("expected name to be generated")
	}
	if name != "manifest" {
		t.Errorf("expected name to take the form of the current directory, got %s", name)
	}
}

func TestLoad(t *testing.T) {
	testcases := []string{"", "duffle.toml", "duffle.json", "duffle.yaml"}

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			m, err := Load(tc, "testdata")
			if err != nil {
				t.Fatal(err)
			}

			if m == nil {
				t.Fatal("manifest should not be nil")
			}

			wantName := "testbundle"
			if m.Name != wantName {
				t.Errorf("expected Name to be %q but got %q", wantName, m.Name)
			}

			if len(m.Components) != 1 {
				t.Fatalf("expected 1 component but got %d", len(m.Components))
			}

			if _, ok := m.Components["cnab"]; !ok {
				t.Errorf("expected a component named cnab but got %v", m.Components)
			}
		})
	}
}
