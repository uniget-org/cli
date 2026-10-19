package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMetadataFromMissingFile(t *testing.T) {
	t.Setenv("UNIGET_IGNORE_METADATA_SIGNATURE", "1")

	config := NewDefaultConfig()
	tools, err := config.LoadMetadata(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if tools != nil {
		t.Errorf("expected no tools, got %+v", tools)
	}
}

func TestLoadMetadata(t *testing.T) {
	t.Setenv("UNIGET_IGNORE_METADATA_SIGNATURE", "1")

	filename := filepath.Join(t.TempDir(), "metadata.json")
	if err := os.WriteFile(filename, []byte(`{"tools":[{"name":"mytool"}]}`), 0o600); err != nil {
		t.Fatalf("failed to write metadata file: %v", err)
	}

	config := NewDefaultConfig()
	tools, err := config.LoadMetadata(filename)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tools == nil {
		t.Fatalf("expected tools, got nil")
	}
	if len(tools.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools.Tools))
	}
	if tools.Tools[0].Name != "mytool" {
		t.Errorf("expected tool name mytool, got %s", tools.Tools[0].Name)
	}
	if tools.Tools[0].Binary != "${target}/bin/mytool" {
		t.Errorf("expected default binary, got %s", tools.Tools[0].Binary)
	}
}
