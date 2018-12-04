package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListKeys(t *testing.T) {
	out := bytes.NewBuffer(nil)
	ring := "../../pkg/signature/testdata/public.gpg"
	if err := listKeys(out, true, ring); err != nil {
		t.Fatal(err)
	}

	is := assert.New(t)
	lines := strings.Split(out.String(), "\n")
	is.Len(lines, 3)
	is.Contains(out.String(), "test2@example.com")

	out.Reset()
	if err := listKeys(out, false, ring); err != nil {
		t.Fatal(err)
	}
	lines = strings.Split(out.String(), "\n")
	is.Len(lines, 4)
	is.Contains(out.String(), "test2@example.com")
	is.Contains(out.String(), "FINGERPRINT")
}
