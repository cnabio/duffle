package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/claim"
)

func TestUpgradePersistsClaim(t *testing.T) {
	out := ioutil.Discard

	tmpDuffleHome, err := setupTmpHomeWithCredentials()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpDuffleHome)

	// 😱 global variables 😱
	duffleHome = tmpDuffleHome

	// Store a dummy claim
	instClaim, err := claim.New("foo")
	instClaim.Bundle = &bundle.Bundle{
		Name:    "bar",
		Version: "0.1.0",
		InvocationImages: []bundle.InvocationImage{
			{
				bundle.BaseImage{Image: "foo/bar:0.1.0", ImageType: "docker"},
			},
		},
	}
	if err != nil {
		t.Fatal(err)
	}
	err = claimStorage().Store(*instClaim)
	if err != nil {
		t.Fatal(err)
	}

	// Upgrade ALL THE THINGS!
	up := &upgradeCmd{
		name:   instClaim.Name,
		driver: "debug",
	}
	up.out = out
	err = up.run()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that the updated claim was persisted
	upClaim, err := claimStorage().Read(instClaim.Name)
	if err != nil {
		t.Fatal(err)
	}
	want := claim.ActionUpgrade
	got := upClaim.Result.Action
	if got != want {
		t.Fatalf("expected the last action on the upClaim to be %q but it was %q", want, got)
	}
}

// setup a temporary duffle home with dummy credentials
func setupTmpHomeWithCredentials() (string, error) {
	tmpDuffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		return tmpDuffleHome, err
	}

	credentialsDir := filepath.Join(tmpDuffleHome, "credentials")
	if err := os.MkdirAll(filepath.Join(tmpDuffleHome, "credentials"), 0755); err != nil {
		return tmpDuffleHome, err
	}
	if err := setupCredentialsDir(credentialsDir); err != nil {
		return tmpDuffleHome, err
	}

	return tmpDuffleHome, nil
}
