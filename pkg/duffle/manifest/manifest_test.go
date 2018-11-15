package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	m := New()
	// Testing to make sure maps are initialized
	is := assert.New(t)
	is.Len(m.Components, 0)
	is.Len(m.Parameters, 0)
	is.Len(m.Credentials, 0)

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
