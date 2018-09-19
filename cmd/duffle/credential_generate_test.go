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

	for _, cred := range creds.Credentials {
		if cred.Name == "first" {
			is.Equal(cred.Destination.EnvVar, credlocs["first"].EnvironmentVariable)
			is.Equal(cred.Source.Value, "EMPTY")
		} else if cred.Name == "second" {
			is.Equal(cred.Destination.EnvVar, credlocs["second"].EnvironmentVariable)
			is.Equal(cred.Destination.Path, credlocs["second"].Path)
		} else {
			t.Fatalf("unexpected credential %s", cred.Name)
		}
	}
}
