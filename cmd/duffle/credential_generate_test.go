package main

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle"

	"github.com/stretchr/testify/assert"
)

func TestGenCredentialSet(t *testing.T) {
	name := "zed"
	credlocs := map[string]bundle.Credential{
		"first": {
			Location: bundle.Location{
				EnvironmentVariable: "FIRST_VAR",
			},
		},
		"second": {
			Location: bundle.Location{
				EnvironmentVariable: "SECOND_VAR",
				Path:                "/second/path",
			},
		},
		"third": {
			Location: bundle.Location{
				Path: "/third/path",
			},
		},
	}
	is := assert.New(t)
	creds, err := genCredentialSet(name, credlocs, genEmptyCredentials)
	is.NoError(err)
	is.Equal(creds.Name, name)
	is.Len(creds.Credentials, 3)

	found := map[string]bool{"first": false, "second": false, "third": false}

	var assignmentOrder []string
	for _, cred := range creds.Credentials {
		assignmentOrder = append(assignmentOrder, cred.Name)
		found[cred.Name] = true
		is.Equal(cred.Source.Value, "EMPTY")
	}

	is.Equal([]string{"first", "second", "third"}, assignmentOrder)

	is.Len(found, 3)
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
