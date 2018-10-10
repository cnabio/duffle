package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/deis/duffle/pkg/version"
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
