package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deislabs/cnab-go/claim"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
)

func TestRunClaimShow(t *testing.T) {
	var buf bytes.Buffer

	t.Run("happy path", func(t *testing.T) {
		csc := claimsShowCmd{
			Name:       "myclaim",
			OnlyBundle: false,
			Storage:    mockClaimStore(),
		}
		expectedClaim := claim.Claim{
			Name: "myclaim",
			Bundle: &bundle.Bundle{
				Name:    "mybundle",
				Version: "0.1.2",
				Outputs: map[string]bundle.Output{
					"some-output": {
						Path: "/some-output-path",
					},
				},
			},
			Outputs: map[string]interface{}{
				"some-output": "some output contents",
			},
		}
		csc.Storage.Store(expectedClaim)
		err := csc.runClaimShow(&buf)
		assert.NoError(t, err)

		var got claim.Claim
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("Attempted to unmarshal claim, but got %s", err)
		}
		assert.Equal(t, expectedClaim, got)
	})

	t.Run("error case: when storage returns error", func(t *testing.T) {
		csc := claimsShowCmd{
			Name:       "unknownclaim",
			OnlyBundle: false,
			Storage:    mockClaimStore(),
		}

		err := csc.runClaimShow(&buf)
		assert.EqualError(t, err, "not found")
	})
}

// happy path: when --bundle is specified
func TestRunClaimShow_Bundle(t *testing.T) {
	var buf bytes.Buffer

	csc := claimsShowCmd{
		Name:       "myclaim",
		OnlyBundle: true,
		Storage:    mockClaimStore(),
	}
	expectedClaim := claim.Claim{
		Name: "myclaim",
		Bundle: &bundle.Bundle{
			Name:    "mybundle",
			Version: "0.1.2",
			Outputs: map[string]bundle.Output{
				"some-output": {
					Path: "/some-output-path",
				},
			},
		},
		Outputs: map[string]interface{}{
			"some-output": "some output contents",
		},
	}
	csc.Storage.Store(expectedClaim)
	err := csc.runClaimShow(&buf)
	assert.NoError(t, err)

	var got bundle.Bundle
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("Attempted to unmarshal bundle, but got %s", err)
	}
	assert.Equal(t, *expectedClaim.Bundle, got)
}

func TestRunClaimShow_Output(t *testing.T) {
	var buf bytes.Buffer
	expectedClaim := claim.Claim{
		Name: "myclaim",
		Bundle: &bundle.Bundle{
			Name:    "mybundle",
			Version: "0.1.2",
			Outputs: map[string]bundle.Output{
				"some-output": {
					Path: "/some-output-path",
				},
			},
		},
		Outputs: map[string]interface{}{
			"some-output": "some output contents",
		},
	}

	t.Run("happy path", func(t *testing.T) {
		csc := claimsShowCmd{
			Name:    "myclaim",
			Output:  "some-output",
			Storage: mockClaimStore(),
		}
		csc.Storage.Store(expectedClaim)

		err := csc.runClaimShow(&buf)
		assert.NoError(t, err)
		assert.Equal(t, expectedClaim.Outputs["some-output"], buf.String())
	})

	t.Run("error case: when --output name is an unknown output", func(t *testing.T) {
		csc := claimsShowCmd{
			Name:    "myclaim",
			Output:  "not-an-output",
			Storage: mockClaimStore(),
		}
		csc.Storage.Store(expectedClaim)

		err := csc.runClaimShow(&buf)
		assert.EqualError(t, err, "unknown output name: not-an-output")
	})

}

//error case: when --bundle and --output are both specified
func TestRunClaimShow_InvalidFlags(t *testing.T) {
	csc := claimsShowCmd{
		Name:       "myclaim",
		Output:     "some-output",
		OnlyBundle: true,
		Storage:    mockClaimStore(),
	}

	err := csc.validateShowFlags()
	assert.EqualError(t, err, "invalid flags: at most one of --bundle and --output can be specified")
}

func TestDisplayJSON(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	expectedClaim := claim.Claim{
		Name: "myclaim",
		Bundle: &bundle.Bundle{
			Name:    "mybundle",
			Version: "0.1.2",
			Outputs: map[string]bundle.Output{
				"some-output": {
					Path: "/some-output-path",
				},
			},
		},
		Outputs: map[string]interface{}{
			"some-output": "some output contents",
		},
	}

	t.Run("happy path", func(t *testing.T) {
		err := displayAsJSON(buf, expectedClaim)
		assert.NoError(t, err)

		var got claim.Claim
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("Attempted to unmarshal claim, but got %s", err)
		}
		assert.Equal(t, expectedClaim, got)
	})

	t.Run("error case: can't marshal claim", func(t *testing.T) {
		err := displayAsJSON(buf, func() {})
		assert.EqualError(t, err, "json: unsupported type: func()")
	})

	t.Run("error case: can't write to output", func(t *testing.T) {
		err := displayAsJSON(mockWriter{err: errors.New("failed to write")}, expectedClaim)
		assert.EqualError(t, err, "failed to write")
	})
}

type mockWriter struct {
	err error
}

func (mw mockWriter) Write([]byte) (int, error) {
	return 0, mw.err
}
