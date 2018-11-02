package main

import (
	"os"
	"testing"

	"github.com/deis/duffle/pkg/reference"
	"github.com/deis/duffle/pkg/repo"
)

func TestPullBundle(t *testing.T) {
	testHome := CreateTestHome(t)

	tests := []struct {
		Name      string
		Reference reference.NamedTagged
	}{
		{
			Name: "helloazure",
			// discard the digest; we don't care about that because it may change over time
			Reference: &NamedRepository{"hub.cnlabs.io/helloazure", "0.1.0", ""},
		},
		{
			Name: "namespaced helloazure",
			// discard the digest; we don't care about that because it may change over time
			Reference: &NamedRepository{"hub.cnlabs.io/library/helloazure", "0.1.0", ""},
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			filePath, err := pullBundle(tc.Reference.String(), true)
			if err != nil {
				t.Error(err)
			}
			defer os.Remove(filePath)

			// check that the repository was recorded in repositories.json
			index, err := repo.LoadIndex(testHome.Repositories())
			if err != nil {
				t.Errorf("cannot create or open %s: %v", testHome.Repositories(), err)
			}

			if !index.Has(tc.Reference.Name(), tc.Reference.Tag()) {
				t.Errorf("could not find entry in %s with name %s and tag %s", testHome.Repositories(), tc.Reference.Name(), tc.Reference.Tag())
			}
		})
	}
}
