package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testMetadata = `{
	"revision": "test-revision",
	"tools": [
		{
			"name": "jq",
			"version": "1.7.1",
			"binary": "jq",
			"description": "A lightweight and flexible command-line JSON processor",
			"tags": ["utility", "json"],
			"runtime_dependencies": ["baselayout"],
			"messages": {
				"internals": "Internals message for {{ .Name }}",
				"usage": "Usage message for {{ .Name }}",
				"update": "Update message for {{ .Name }}"
			}
		},
		{
			"name": "baselayout",
			"version": "3.5.1",
			"binary": "false",
			"description": "baselayout for testing"
		}
	]
}`

// resetUnigetEnv makes tests hermetic by neutralizing UNIGET_* environment
// variables from the invoking shell. Without this, UNIGET_USER=true would
// switch the configuration to user mode and override the prefix with $HOME.
func resetUnigetEnv(t *testing.T) {
	t.Helper()
	t.Setenv("UNIGET_USER", "false")
}

// seedMetadata points the configuration to a temporary prefix containing a
// metadata file for the test fixture. Loading it does not require network
// access because UNIGET_IGNORE_METADATA_SIGNATURE bypasses signature
// verification and the file existing bypasses the download.
func seedMetadata(t *testing.T) string {
	t.Helper()
	resetUnigetEnv(t)

	prefix := t.TempDir()
	t.Setenv("UNIGET_PREFIX", prefix)
	t.Setenv("UNIGET_IGNORE_METADATA_SIGNATURE", "1")

	metadataFile := filepath.Join(prefix, "var", "cache", "uniget", "metadata.json")
	if err := os.MkdirAll(filepath.Dir(metadataFile), 0o750); err != nil {
		t.Fatalf("failed to create metadata directory: %v", err)
	}
	if err := os.WriteFile(metadataFile, []byte(testMetadata), 0o600); err != nil {
		t.Fatalf("failed to write metadata file: %v", err)
	}

	return prefix
}

// installToolFixture fakes an installed jq tool: The binary is present and a
// marker file for the given version exists in the cache directory.
func installToolFixture(t *testing.T, prefix string, markerVersion string) {
	t.Helper()

	binaryFile := filepath.Join(prefix, "usr", "local", "bin", "jq")
	if err := os.MkdirAll(filepath.Dir(binaryFile), 0o750); err != nil {
		t.Fatalf("failed to create binary directory: %v", err)
	}
	if err := os.WriteFile(binaryFile, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil { // #nosec G306 -- Test fixture needs executable bit
		t.Fatalf("failed to write binary: %v", err)
	}

	markerDirectory := filepath.Join(prefix, "var", "cache", "uniget", "jq")
	if err := os.MkdirAll(markerDirectory, 0o750); err != nil {
		t.Fatalf("failed to create marker file directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(markerDirectory, markerVersion), nil, 0o600); err != nil {
		t.Fatalf("failed to write marker file: %v", err)
	}
}

func TestTags(t *testing.T) {
	seedMetadata(t)

	out, err := runCobraTest(t, nil, "tags")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	for _, expected := range []string{"utility", "json"} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain <%s>, received <%s>", expected, out)
		}
	}
}

func TestSearch(t *testing.T) {
	seedMetadata(t)

	// Warning: Flags are bound to global variables which are not reset
	// between cobra executions. Cases are ordered from least to most
	// polluting and always pass --output explicitly.

	tt := []cobraTest{
		{
			name:        "table output",
			args:        []string{"search", "--output", "table", "jq"},
			expectOut:   "jq",
			outContains: true,
		},
		{
			name:        "search in description",
			args:        []string{"search", "--output", "table", "--only-description", "JSON"},
			expectOut:   "jq",
			outContains: true,
		},
		{
			name:        "search in tags",
			args:        []string{"search", "--output", "table", "--only-tags", "utility"},
			expectOut:   "jq",
			outContains: true,
		},
		{
			name:        "no match",
			args:        []string{"search", "--output", "table", "doesnotexist"},
			expectOut:   "No tools found",
			outContains: true,
		},
		{
			name:      "invalid output format",
			args:      []string{"search", "--output", "bogus", "jq"},
			expectErr: errors.New("error: output format bogus not supported"),
		},
		{
			name:      "conflicting only flags",
			args:      []string{"search", "--only-names", "--only-tags", "jq"},
			expectErr: errors.New("error: Can only process one of only-names, only-description, only-tags and only-deps at the same time"),
		},
	}
	runCobraTests(t, tt)

	// Name and json output are printed to stdout by the command.
	for _, args := range [][]string{
		{"search", "--output", "name", "jq"},
		{"search", "--output", "json", "jq"},
	} {
		_, stdout, err := runCobraTestCaptureStdout(t, args...)
		if err != nil {
			t.Errorf("command %v returned unexpected error: %v", args, err)
		}
		if !strings.Contains(stdout, "jq") {
			t.Errorf("expected stdout of %v to contain <jq>, received <%s>", args, stdout)
		}
	}
}

func TestList(t *testing.T) {
	seedMetadata(t)

	// Warning: Flags are bound to global variables which are not reset
	// between cobra executions. Cases are ordered from least to most
	// polluting and always pass --output explicitly.

	tt := []cobraTest{
		{
			name:        "pretty output",
			args:        []string{"list", "--output", "pretty"},
			expectOut:   "jq",
			outContains: true,
		},
		{
			name:        "yaml output",
			args:        []string{"list", "--output", "yaml"},
			expectOut:   "name: jq",
			outContains: true,
		},
		{
			name:      "invalid output format",
			args:      []string{"list", "--output", "bogus"},
			expectErr: errors.New("invalid output format: bogus"),
		},
	}
	runCobraTests(t, tt)

	// JSON output is printed to stdout by the command.
	_, stdout, err := runCobraTestCaptureStdout(t, "list", "--output", "json")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "jq") {
		t.Errorf("expected stdout to contain <jq>, received <%s>", stdout)
	}
}

func TestListInstalled(t *testing.T) {
	prefix := seedMetadata(t)
	installToolFixture(t, prefix, "1.7.1")

	out, err := runCobraTest(t, nil, "list", "--output", "pretty", "--installed")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "jq") {
		t.Errorf("expected output to contain <jq>, received <%s>", out)
	}
}

func TestListUpgradable(t *testing.T) {
	prefix := seedMetadata(t)
	installToolFixture(t, prefix, "1.0.0")

	out, err := runCobraTest(t, nil, "list", "--output", "pretty", "--upgradable")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "jq") {
		t.Errorf("expected output to contain <jq>, received <%s>", out)
	}
}

func TestDescribe(t *testing.T) {
	seedMetadata(t)

	tt := []cobraTest{
		{
			name:        "pretty output",
			args:        []string{"describe", "--output", "pretty", "jq"},
			expectOut:   "Name: jq",
			outContains: true,
		},
		{
			name:        "yaml output",
			args:        []string{"describe", "--output", "yaml", "jq"},
			expectOut:   "name: jq",
			outContains: true,
		},
		{
			name:      "invalid output format",
			args:      []string{"describe", "--output", "bogus", "jq"},
			expectErr: errors.New("invalid output format: bogus"),
		},
		{
			name:      "unknown tool",
			args:      []string{"describe", "doesnotexist"},
			expectErr: errors.New("error getting tool doesnotexist: Tool named doesnotexist not found"),
		},
	}
	runCobraTests(t, tt)

	// JSON output is printed to stdout by the command.
	_, stdout, err := runCobraTestCaptureStdout(t, "describe", "--output", "json", "jq")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `"name":"jq"`) {
		t.Errorf("expected stdout to contain <\"name\":\"jq\">, received <%s>", stdout)
	}
}

func TestMessage(t *testing.T) {
	seedMetadata(t)

	out, err := runCobraTest(t, nil, "message", "jq")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	for _, expected := range []string{
		"Internals message for jq",
		"Usage message for jq",
		"Update message for jq",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected output to contain <%s>, received <%s>", expected, out)
		}
	}

	// Messages are printed to stdout by these commands.
	_, stdout, err := runCobraTestCaptureStdout(t, "message", "--list", "jq")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	for _, expected := range []string{"Internals", "Usage", "Update"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("expected stdout to contain <%s>, received <%s>", expected, stdout)
		}
	}

	_, stdout, err = runCobraTestCaptureStdout(t, "message", "--find", "jq")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "jq") {
		t.Errorf("expected stdout to contain <jq>, received <%s>", stdout)
	}

	// --find also works without a tool name. It used to panic.
	_, stdout, err = runCobraTestCaptureStdout(t, "message", "--find")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "jq") {
		t.Errorf("expected stdout to contain <jq>, received <%s>", stdout)
	}

	// --list without a tool name returns an error.
	_, err = runCobraTest(t, nil, "message", "--list")
	if err == nil {
		t.Error("expected error for missing tool name")
	} else if !strings.Contains(err.Error(), "message list requires a tool name") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}

	// No messages for baselayout
	_, stdout, err = runCobraTestCaptureStdout(t, "message", "baselayout")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, received <%s>", stdout)
	}

	_, err = runCobraTest(t, nil, "message", "doesnotexist")
	if err == nil {
		t.Error("expected error for unknown tool")
	} else if !strings.Contains(err.Error(), "failed to get tool") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}
}

func TestVersion(t *testing.T) {
	prefix := seedMetadata(t)

	// Unknown tool
	_, err := runCobraTest(t, nil, "version", "doesnotexist")
	if err == nil {
		t.Error("expected error for unknown tool")
	} else if !strings.Contains(err.Error(), "failed to get tool") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}

	// Not installed
	_, err = runCobraTest(t, nil, "version", "baselayout")
	if err == nil {
		t.Error("expected error for uninstalled tool")
	} else if !strings.Contains(err.Error(), "tool baselayout is not installed") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}

	// Installed but without version check. The version is printed to stdout.
	installToolFixture(t, prefix, "1.7.1")
	_, stdout, err := runCobraTestCaptureStdout(t, "version", "jq")
	if err != nil {
		t.Errorf("returned unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "1.7.1") {
		t.Errorf("expected stdout to contain <1.7.1>, received <%s>", stdout)
	}
}

func TestHealthcheck(t *testing.T) {
	prefix := seedMetadata(t)

	// Unknown tool
	_, err := runCobraTest(t, nil, "healthcheck", "doesnotexist")
	if err == nil {
		t.Error("expected error for unknown tool")
	} else if !strings.Contains(err.Error(), "error getting tool doesnotexist") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}

	// Tool without binary and marker file
	_, err = runCobraTest(t, nil, "healthcheck", "baselayout")
	if err == nil {
		t.Error("expected error for uninstalled tool")
	} else if !strings.Contains(err.Error(), "healthcheck failed for baselayout") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}

	// Installed tool with binary and marker file
	installToolFixture(t, prefix, "1.7.1")
	if _, err := runCobraTest(t, nil, "healthcheck", "jq"); err != nil {
		t.Errorf("returned unexpected error: %v", err)
	}
}

func TestDeps(t *testing.T) {
	seedMetadata(t)

	// The dependency tree is printed to stdout by the command.
	_, stdout, err := runCobraTestCaptureStdout(t, "deps", "tree", "jq")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	for _, expected := range []string{"jq", "baselayout"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("expected stdout to contain <%s>, received <%s>", expected, stdout)
		}
	}

	_, stdout, err = runCobraTestCaptureStdout(t, "deps", "tree")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	for _, expected := range []string{"jq", "baselayout"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("expected stdout to contain <%s>, received <%s>", expected, stdout)
		}
	}

	_, stdout, err = runCobraTestCaptureStdout(t, "deps", "who-uses", "baselayout")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "jq") {
		t.Errorf("expected stdout to contain <jq>, received <%s>", stdout)
	}

	// Nothing depends on jq
	_, stdout, err = runCobraTestCaptureStdout(t, "deps", "who-uses", "jq")
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, received <%s>", stdout)
	}

	_, err = runCobraTest(t, nil, "deps", "tree", "doesnotexist")
	if err == nil {
		t.Error("expected error for unknown tool")
	} else if !strings.Contains(err.Error(), "error getting tool doesnotexist") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}
}

func TestBuildVersion(t *testing.T) {
	if buildVersion() == "" {
		t.Error("expected non-empty version")
	}
}
