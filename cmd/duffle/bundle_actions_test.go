package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestBundleActions(t *testing.T) {
	out := bytes.NewBuffer(nil)
	cmd := &bundleActionsCmd{
		out:          out,
		bundleIsFile: true,
		bundle:       "./testdata/testbundle/actions-test-bundle.json",
		insecure:     true,
	}

	if err := cmd.run(); err != nil {
		t.Fatal(err)
	}

	result := out.String()

	newLines := strings.Count(result, "\n")

	// the bundle contains 3 actions, plus the table header
	// so the output should have 4 lines
	if newLines != 4 {
		t.Errorf("Expected output to have 4 lines. It had %v", newLines)
	}

	if !strings.Contains(result, "dry-run") {
		t.Errorf("Expected actions to contain %s but did not", "dry-run")
	}
	if !strings.Contains(result, "migrate") {
		t.Errorf("Expected actions to contain %s but did not", "migrate")
	}

	if !strings.Contains(result, "status") {
		t.Errorf("Expected actions to contain %s but did not", "status")
	}
}
