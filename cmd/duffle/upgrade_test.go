package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/claim"
)

func TestUpgradePersistsClaim(t *testing.T) {
	out := ioutil.Discard

	// Setup a temporary duffle home with dummy credentials
	tmpDuffleHome, err := ioutil.TempDir("", "dufflehome")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpDuffleHome)

	credentialsDir := filepath.Join(tmpDuffleHome, "credentials")
	if err := os.MkdirAll(filepath.Join(tmpDuffleHome, "credentials"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := setupCredentialsDir(credentialsDir); err != nil {
		t.Fatal(err)
	}

	// 😱 global variables 😱
	duffleHome = tmpDuffleHome
	upgradeDriver = "debug"

	// Store a dummy claim
	instClaim, err := claim.New("foo")
	instClaim.Bundle = &bundle.Bundle{
		Name:    "bar",
		Version: "0.1.0",
		InvocationImages: []bundle.InvocationImage{
			{Image: "foo/bar:0.1.0", ImageType: "docker"},
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
		name: instClaim.Name,
	}
	up.Out = out
	err = up.upgrade("", "")
	if err != nil {
		t.Fatal(err)
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
