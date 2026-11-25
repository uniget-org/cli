package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testLoadToolsString = `{
	"tools": [
		{
			"name":"foo",
			"version":"1.0.0",
			"runtime_dependencies": [
				"bar"
			],
			"tags": [
				"baz",
				"blarg"
			],
			"repository": "https://foo.bar",
			"description": "Foo"
		},
		{
			"name":"bar",
			"version":"2.0.0",
			"tags": [
				"baz",
				"blubb"
			]
		}
	]
}`

func TestLoadFromBytes(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testLoadToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	if len(tools.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools.Tools))
	}

	foo, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool foo: %s\n", err)
	}

	if foo.Name != "foo" {
		t.Errorf("Expected foo, got %s", foo.Name)
	}
	if foo.Version != "1.0.0" {
		t.Errorf("Expected 1.0.0, got %s", foo.Version)
	}
	if len(foo.RuntimeDependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(foo.RuntimeDependencies))
	}
	if foo.RuntimeDependencies[0] != "bar" {
		t.Errorf("Expected bar, got %s", foo.RuntimeDependencies[0])
	}
}

func TestLoadFromBytesRejectsUnsafePathComponents(t *testing.T) {
	tests := []struct {
		name    string
		tool    string
		version string
	}{
		{name: "parent traversal name", tool: "../escape", version: "1.0.0"},
		{name: "absolute name", tool: "/escape", version: "1.0.0"},
		{name: "nested name", tool: "sub/dir", version: "1.0.0"},
		{name: "dot name", tool: ".", version: "1.0.0"},
		{name: "empty name", tool: "", version: "1.0.0"},
		{name: "parent traversal version", tool: "jq", version: "../../escape"},
		{name: "absolute version", tool: "jq", version: "/escape"},
		{name: "nested version", tool: "jq", version: "sub/dir"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := fmt.Sprintf(`{"tools":[{"name":%q,"version":%q}]}`, tt.tool, tt.version)
			if _, err := LoadFromBytes([]byte(data)); err == nil {
				t.Fatalf("expected error for tool %q version %q, got nil", tt.tool, tt.version)
			}
		})
	}
}

func TestLoadFromBytesAcceptsSafePathComponents(t *testing.T) {
	data := `{"tools":[{"name":"my-tool_1.2","version":"1.2.3-rc1+build.5"}]}`
	tools, err := LoadFromBytes([]byte(data))
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(tools.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools.Tools))
	}
}

func TestLoadFromBytesRejectsControlCharacterCheck(t *testing.T) {
	data := `{
		"tools": [
			{
				"name": "evil",
				"version": "1.0.0",
				"check": "echo 1.2.3\nrm -rf /"
			}
		]
	}`

	_, err := LoadFromBytes([]byte(data))
	if err == nil {
		t.Fatalf("expected error for check with control characters, got nil")
	}
	if !strings.Contains(err.Error(), "control characters") {
		t.Errorf("expected error to mention control characters, got %s", err)
	}
}

func TestValidateToolCheck(t *testing.T) {
	tests := []struct {
		name    string
		check   string
		wantErr bool
	}{
		{name: "empty check", check: ""},
		{name: "simple command", check: "jq --version | cut -d- -f2"},
		{name: "variable substitution", check: "${binary} --version 2>/dev/null | cut -d+ -f1"},
		{name: "newline", check: "echo 1.2.3\nrm -rf /", wantErr: true},
		{name: "carriage return", check: "echo 1.2.3\r", wantErr: true},
		{name: "nul byte", check: "echo \x00 1.2.3", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToolCheck(tt.check)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "tools.json")
	if err := os.WriteFile(filePath, []byte(testLoadToolsString), 0o600); err != nil {
		t.Fatalf("failed to write tools file: %s", err)
	}

	tools, err := LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(tools.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools.Tools))
	}

	_, err = LoadFromFile(filepath.Join(tmpDir, "nonexistent.json"))
	if err == nil {
		t.Errorf("expected error for missing file, got nil")
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("precondition: expected file to exist: %v", err)
	}

	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("not json"), 0o600); err != nil {
		t.Fatalf("failed to write invalid file: %s", err)
	}
	_, err = LoadFromFile(invalidPath)
	if err == nil || !strings.Contains(err.Error(), "error loading data") {
		t.Errorf("expected json error, got %v", err)
	}
}

func TestLoadFromReaderFillsDefaultBinary(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testLoadToolsString))
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	for _, tool := range tools.Tools {
		switch tool.Name {
		case "foo":
			if tool.Binary != "${target}/bin/foo" {
				t.Errorf("expected explicit binary to be kept, got %q", tool.Binary)
			}
		case "bar":
			if tool.Binary != "${target}/bin/bar" {
				t.Errorf("expected default binary for %q, got %q", tool.Name, tool.Binary)
			}
		}
	}

	_, err = LoadFromBytes([]byte("]not json["))
	if err == nil {
		t.Errorf("expected error for invalid json, got nil")
	}
}

func TestLoadManifestFromBytes(t *testing.T) {
	manifest := []byte("name: jq\nversion: 1.0.0\n")

	tool, err := LoadManifestFromBytes(manifest)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if tool.Name != "jq" {
		t.Errorf("expected tool jq, got %s", tool.Name)
	}
	if tool.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", tool.Version)
	}

	_, err = LoadManifestFromBytes([]byte("[unclosed"))
	if err == nil || !strings.Contains(err.Error(), "unable to load YAML") {
		t.Errorf("expected yaml error, got %v", err)
	}
}

func TestLoadManifestFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "manifest.yaml")
	if err := os.WriteFile(filePath, []byte("name: jq\nversion: 1.0.0\n"), 0o600); err != nil {
		t.Fatalf("failed to write manifest: %s", err)
	}

	tool, err := LoadManifestFromFile(filePath)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if tool.Name != "jq" {
		t.Errorf("expected tool jq, got %s", tool.Name)
	}

	_, err = LoadManifestFromFile(filepath.Join(tmpDir, "nonexistent.yaml"))
	if err == nil {
		t.Errorf("expected error for missing file, got nil")
	}
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("precondition: expected file to exist: %v", err)
	}
}
