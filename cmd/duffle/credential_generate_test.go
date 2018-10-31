package main

import (
	"testing"

	"github.com/deis/duffle/pkg/bundle"

	"github.com/stretchr/testify/assert"
)

func TestGenCredentialSet(t *testing.T) {
	name := "zed"
	credlocs := map[string]bundle.CredentialLocation{
		"first": {
			EnvironmentVariable: "FIRST_VAR",
		},
		"second": {
			EnvironmentVariable: "SECOND_VAR",
			Path:                "/second/path",
		},
	}
	creds := genCredentialSet(name, credlocs)
	is := assert.New(t)
	is.Equal(creds.Name, name)
	is.Len(creds.Credentials, 2)

	found := map[string]bool{"first": false, "second": false}

	for _, cred := range creds.Credentials {
		found[cred.Name] = true
		is.Equal(cred.Source.Value, "EMPTY")
	}

	is.Len(found, 2)
	for k, v := range found {
		is.True(v, "%q not found", k)
	}
}
