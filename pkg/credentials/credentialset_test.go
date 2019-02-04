package credentials

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/scothis/ruffle/pkg/bundle"

	"github.com/stretchr/testify/assert"
)

func TestCredentialSet(t *testing.T) {
	is := assert.New(t)
	if err := os.Setenv("TEST_USE_VAR", "kakapu"); err != nil {
		t.Fatal("could not setup env")
	}
	defer os.Unsetenv("TEST_USE_VAR")

	goos := "unix"
	if runtime.GOOS == "windows" {
		goos = runtime.GOOS
	}
	credset, err := Load(fmt.Sprintf("testdata/staging-%s.yaml", goos))
	is.NoError(err)

	results, err := credset.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	count := 5
	is.Len(results, count, "Expected %d credentials", count)

	for _, tt := range []struct {
		name   string
		key    string
		expect string
		path   string
	}{
		{name: "run_program", key: "TEST_RUN_PROGRAM", expect: "wildebeest"},
		{name: "use_var", key: "TEST_USE_VAR", expect: "kakapu"},
		{name: "read_file", key: "TEST_READ_FILE", expect: "serval"},
		{name: "fallthrough", key: "TEST_FALLTHROUGH", expect: "quokka", path: "/animals/quokka.txt"},
		{name: "plain_value", key: "TEST_PLAIN_VALUE", expect: "cassowary"},
	} {
		dest, ok := results[tt.name]
		is.True(ok)
		is.Equal(tt.expect, strings.TrimSpace(dest))
	}
}

func TestCredentialSet_Expand(t *testing.T) {
	b := &bundle.Bundle{
		Name: "knapsack",
		Credentials: map[string]bundle.Location{
			"first": {
				EnvironmentVariable: "FIRST_VAR",
			},
			"second": {
				Path: "/second/path",
			},
			"third": {
				EnvironmentVariable: "/THIRD_VAR",
				Path:                "/third/path",
			},
		},
	}
	cs := Set{
		"first":  "first",
		"second": "second",
		"third":  "third",
	}

	env, path, err := cs.Expand(b)
	is := assert.New(t)
	is.NoError(err)
	for k, v := range b.Credentials {
		if v.EnvironmentVariable != "" {
			is.Equal(env[v.EnvironmentVariable], cs[k])
		}
		if v.Path != "" {
			is.Equal(path[v.Path], cs[k])
		}
	}
}
