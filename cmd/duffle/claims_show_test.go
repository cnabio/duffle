package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/deislabs/cnab-go/claim"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
)

func TestDisplayClaim(t *testing.T) {
	var buf bytes.Buffer
	storage := mockClaimStore()

	storage.Store(claim.Claim{
		Name: "myclaim",
		Bundle: &bundle.Bundle{
			Name:    "mybundle",
			Version: "0.1.2",
		},
	})

	displayClaim("myclaim", &buf, storage, false)

	var got claim.Claim
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	is := assert.New(t)
	is.Equal("myclaim", got.Name)
	is.Equal("mybundle", got.Bundle.Name)
}

func TestDisplayClaim_Bundle(t *testing.T) {
	var buf bytes.Buffer
	storage := mockClaimStore()

	storage.Store(claim.Claim{
		Name: "myclaim",
		Bundle: &bundle.Bundle{
			Name:    "mybundle",
			Version: "0.1.2",
		},
	})

	displayClaim("myclaim", &buf, storage, true)

	var got bundle.Bundle
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	is := assert.New(t)
	is.Equal("mybundle", got.Name)
	is.Equal("0.1.2", got.Version)
}
