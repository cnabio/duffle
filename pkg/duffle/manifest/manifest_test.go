package manifest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	m := New()
	// Testing to make sure maps are initialized
	is := assert.New(t)
	is.Len(m.InvocationImages, 0)
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

			if len(m.InvocationImages) != 1 {
				t.Fatalf("expected 1 component but got %d", len(m.InvocationImages))
			}

			if _, ok := m.InvocationImages["cnab"]; !ok {
				t.Errorf("expected a component named cnab but got %v", m.InvocationImages)
			}

			if len(m.Parameters) != 1 {
				t.Fatalf("expected 1 parameter but got %d", len(m.Parameters))
			}

			param, ok := m.Parameters["foo"]
			if !ok {
				t.Errorf("expected a parameter named foo but got %v", m.Parameters)
			}

			if param.DataType != "string" {
				t.Errorf("expected foo parameter to have a type of string but got %v", param.DataType)
			}

			if len(m.Credentials) != 1 {
				t.Fatalf("expected 1 credential but got %d", len(m.Credentials))
			}

			cred, ok := m.Credentials["bar"]
			if !ok {
				t.Errorf("expected a credential named bar but got %v", m.Credentials)
			}

			if cred.Path != "/tmp" {
				t.Errorf("expected foo credential to have a path of /tmp but got %v", cred.Path)
			}

			if len(m.Maintainers) != 1 {
				t.Fatalf("expected 1 maintainer but got %d", len(m.Maintainers))
			}

			maintainer := m.Maintainers[0]
			if maintainer.Name != "sally" {
				t.Errorf("expected maintainer to be sally but got %v", maintainer.Name)
			}
		})
	}
}

func TestInvalidLoad(t *testing.T) {
	testcases := []string{"invalid_duffle.json"}

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			_, err := Load(tc, "testdata")
			if err == nil {
				t.Errorf("expected an error to be thrown")
			}
			if !strings.Contains(err.Error(), "error(s) decoding") {
				t.Errorf("expected err to contain %s but was %s", "error(s) decoding", err.Error())
			}
		})
	}
}

func TestExamples(t *testing.T) {
	testcases := []string{"helloworld/duffle.json"}

	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			_, err := Load(tc, "../../../examples")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
