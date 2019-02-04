package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scothis/ruffle/pkg/bundle"
	"github.com/scothis/ruffle/pkg/version"
)

func TestHelpWrittenToStdout(t *testing.T) {
	// Swap out stdout/stderr just for this test
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	// duffle help
	cmd := newRootCmd(nil)
	cmd.SetArgs([]string{"help"})
	cmd.Execute()

	// copy the output in a separate goroutine so printing can't block indefinitely
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// read our faked stdout
	w.Close()
	gotstdout := <-outC

	// Verify help text went to stdout
	wantHelpText := cmd.UsageString()
	if !strings.Contains(gotstdout, wantHelpText) {
		t.Fatalf("expected help text to be printed to stdout but got %q", gotstdout)
	}
}

func TestLogsWrittenToStdout(t *testing.T) {
	// Swap out stdout/stderr just for this test
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set a temp version
	version.Version = "v1.0.0"

	// duffle version
	cmd := newRootCmd(nil)
	cmd.SetArgs([]string{"version"})
	cmd.Execute()

	// copy the output in a separate goroutine so printing can't block indefinitely
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// read our faked stdout
	w.Close()
	gotstdout := <-outC

	// Verify logs went to stdout
	wantLogs := version.Version
	if !strings.Contains(gotstdout, wantLogs) {
		t.Fatalf("expected logs to be printed to stdout but got %q", gotstdout)
	}
}

func TestErrorsWrittenToStderr(t *testing.T) {
	// Swap out stdout/stderr just for this test
	stderr := os.Stderr
	defer func() {
		os.Stderr = stderr
	}()
	r, w, _ := os.Pipe()
	os.Stderr = w

	// duffle build (without being properly setup triggers an error)
	cmd := newRootCmd(nil)
	cmd.SetArgs([]string{"build"})
	err := cmd.Execute()

	// copy the output in a separate goroutine so printing can't block indefinitely
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// read our faked stderr
	w.Close()
	gotstderr := <-outC

	// Verify error text went to stderr
	wantError := err.Error()
	if !strings.Contains(gotstderr, wantError) {
		t.Fatalf("expected error text to be printed to stderr but got %q", gotstderr)
	}
}

func TestSearchOutputDefault(t *testing.T) {
	t.Skip("under construction")
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	duffleHome = filepath.Join(cwd, "..", "..", "tests", "testdata", "home")

	// duffle search
	cmd := newRootCmd(nil)
	cmd.SetArgs([]string{"search", "hello"})
	err = cmd.Execute()
	if err != nil {
		t.Error(err)
	}

	// copy the output in a separate goroutine so printing can't block indefinitely
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	gotstdout := <-outC

	// Verify stdout shows default (table) listing
	wantOutput := "NAME      \tVERSION\n"
	if !strings.Contains(gotstdout, wantOutput) {
		t.Fatalf("expected default table output, got %q", gotstdout)
	}
}

func TestSearchOutputJSON(t *testing.T) {
	t.Skip("under construction")
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	duffleHome = filepath.Join(cwd, "..", "..", "tests", "testdata", "home")

	// duffle search -o json
	cmd := newRootCmd(nil)
	cmd.SetArgs([]string{"search", "hello", "-o", "json"})
	err = cmd.Execute()
	if err != nil {
		t.Error(err)
	}

	// copy the output in a separate goroutine so printing can't block indefinitely
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	gotstdout := <-outC

	// Verify stdout shows json listing
	var bundleList []*bundle.Bundle
	err = json.Unmarshal([]byte(gotstdout), &bundleList)
	if err != nil {
		t.Fatalf("expected json output, got %q", gotstdout)
	}
}

func TestSearchOutputInvalid(t *testing.T) {
	t.Skip("under construction")
	cmd := newRootCmd(nil)
	cmd.SetArgs([]string{"search", "hello", "-o", "bogus"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error to be returned")
	}
}
