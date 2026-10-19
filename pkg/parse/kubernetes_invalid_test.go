package parse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadKubernetesManifestWithInvalidManifest(t *testing.T) {
	// Previously, invalid manifests caused a panic instead of an error.
	_, err := LoadKubernetesManifest([]byte("this is not a kubernetes manifest"))
	if err == nil {
		t.Fatalf("expected error for invalid manifest, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode kubernetes manifest") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}
}

func TestLoadKubernetesManifestFromFileWithInvalidManifest(t *testing.T) {
	manifestFile := filepath.Join(t.TempDir(), "manifest.yaml")
	if err := os.WriteFile(manifestFile, []byte("invalid {{ manifest"), 0o600); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	_, err := LoadKubernetesManifestFromFile(manifestFile)
	if err == nil {
		t.Fatalf("expected error for invalid manifest, got nil")
	}
	if !strings.Contains(err.Error(), "failed to load kubernetes manifest") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}
}
