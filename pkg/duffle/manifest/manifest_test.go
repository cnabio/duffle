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
	is := assert.New(t)
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
			is.Equal(wantName, m.Name)
			is.Len(m.InvocationImages, 1)
			if _, ok := m.InvocationImages["cnab"]; !ok {
				t.Fatalf("expected a component named cnab but got %v", m.InvocationImages)
			}

			is.Len(m.Parameters, 1)
			param, ok := m.Parameters["foo"]
			is.True(ok, "param should exist")
			is.Equal("string", param.DataType)
			is.Len(m.Credentials, 1)

			cred, ok := m.Credentials["bar"]
			is.True(ok, "expected credential")
			is.Equal("/tmp", cred.Path)
			is.Len(m.Maintainers, 1)

			maintainer := m.Maintainers[0]
			is.Equal("sally", maintainer.Name)

			is.Len(m.Images, 1)
			is.Equal("test:latest", m.Images["test"].Image)

			is.Len(m.Actions, 1)
			is.Equal("says hello", m.Actions["hello"].Description)
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
