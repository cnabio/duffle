package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/deislabs/duffle/pkg/duffle/home"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/credentials"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func CreateTestHome(t *testing.T) home.Home {
	t.Helper()
	tempDir, err := ioutil.TempDir("", "duffle")
	if err != nil {
		t.Fatal(err)
	}
	duffleHome = tempDir
	testHome := home.Home(tempDir)
	dirs := []string{
		testHome.String(),
		testHome.Bundles(),
		testHome.Logs(),
		testHome.Plugins(),
		testHome.Claims(),
		testHome.Credentials(),
	}
	if err := ensureDirectories(dirs); err != nil {
		t.Fatal(err)
	}
	if err := ensureFiles([]string{testHome.Repositories()}); err != nil {
		t.Fatal(err)
	}
	return testHome
}

func TestLoadCredentials(t *testing.T) {
	cred1 := credentials.CredentialSet{
		Name: "first",
		Credentials: []credentials.CredentialStrategy{
			{Name: "knapsack", Source: credentials.Source{Value: "cred1"}},
			{Name: "gym-bag", Source: credentials.Source{Value: "cred1"}},
		},
	}
	cred2 := credentials.CredentialSet{
		Name: "second",
		Credentials: []credentials.CredentialStrategy{
			{Name: "knapsack", Source: credentials.Source{Value: "cred2"}},
			{Name: "haversack", Source: credentials.Source{Value: "cred2"}},
		},
	}
	cred3 := credentials.CredentialSet{
		Name: "third",
		Credentials: []credentials.CredentialStrategy{
			{Name: "haversack", Source: credentials.Source{Value: "cred3"}},
		},
	}

	// The above should generate:
	// -- knapsack: cred2
	// -- havershack: cred3
	// -- gym-bag: cred1

	tmpdir, err := ioutil.TempDir("", "duffle-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	files := make([]string, 3)
	for i, c := range []credentials.CredentialSet{cred1, cred2, cred3} {
		data, err := yaml.Marshal(c)
		t.Log(string(data))
		if err != nil {
			t.Fatal(err)
		}
		fp := filepath.Join(tmpdir, c.Name+".yaml")
		if err := ioutil.WriteFile(fp, data, 0644); err != nil {
			t.Fatal(err)
		}
		files[i] = fp
	}

	bun := bundle.Bundle{
		Name: "test-load-creds",
		Credentials: map[string]bundle.Credential{
			"knapsack": {
				Location: bundle.Location{
					EnvironmentVariable: "KNAP",
				},
			},
			"haversack": {
				Location: bundle.Location{
					EnvironmentVariable: "HAVER",
				},
			},
			"gym-bag": {
				Location: bundle.Location{
					EnvironmentVariable: "GYM",
				},
			},
		},
	}

	is := assert.New(t)
	creds, err := loadCredentials(files, &bun)
	is.NoError(err)
	is.Equal("cred2", creds["knapsack"])
	is.Equal("cred3", creds["haversack"])
	is.Equal("cred1", creds["gym-bag"])
}

func TestFindCreds(t *testing.T) {
	credDir, err := ioutil.TempDir("", "credTest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(credDir)
	is := assert.New(t)

	tests := map[string]struct {
		input            string
		expectedFilePath func() string
	}{
		"no path yaml": {
			input: "creds1",
			expectedFilePath: func() string {
				credPath := filepath.Join(credDir, "creds1.yaml")
				err = ioutil.WriteFile(credPath, []byte("test"), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return credPath
			},
		},
		"no path yml": {
			input: "creds2",
			expectedFilePath: func() string {
				credPath := filepath.Join(credDir, "creds2.yml")
				err = ioutil.WriteFile(credPath, []byte("test"), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return credPath
			},
		},
		"path": {
			input: "testdata/dufflehome/credentials/testing.yaml",
			expectedFilePath: func() string {
				return "testdata/dufflehome/credentials/testing.yaml"
			},
		},
	}

	for name, testCase := range tests {
		is.Equal(testCase.expectedFilePath(), findCreds(credDir, testCase.input), "Fail on test: "+name)
		os.RemoveAll(filepath.Join(credDir, "*"))
	}
}
