package main

import (
	"testing"

	"github.com/deislabs/duffle/pkg/bundle"

	"github.com/stretchr/testify/assert"
)

func TestGenCredentialSet(t *testing.T) {
	name := "zed"
	credlocs := map[string]bundle.Location{
		"first": {
			EnvironmentVariable: "FIRST_VAR",
		},
		"second": {
			EnvironmentVariable: "SECOND_VAR",
			Path:                "/second/path",
		},
	}
	is := assert.New(t)
	creds, err := genCredentialSet(name, credlocs, genEmptyCredentials)
	is.NoError(err)
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

func TestGenCredentialSetBadName(t *testing.T) {
	testcases := []string{
		"period.",
		"forwardslash/",
		"backslash\\",
		"all.of.the/above\\",
	}
	for _, tc := range testcases {
		t.Run(tc, func(t *testing.T) {
			is := assert.New(t)
			_, err := genCredentialSet(tc, nil, genEmptyCredentials)
			is.Error(err)
		})
	}
}
