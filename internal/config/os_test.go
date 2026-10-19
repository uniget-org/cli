package config

import (
	"os"
	"testing"
)

func TestAssertWritableTarget(t *testing.T) {
	clearConfigEnv(t)

	config := NewDefaultConfig(WithPrefix(t.TempDir()))
	config.AssertWritableTarget()
}

func TestAssertLibDirectory(t *testing.T) {
	clearConfigEnv(t)

	config := NewDefaultConfig(WithPrefix(t.TempDir()))
	config.SetGlobalConfig()
	config.AssertLibDirectory()

	if _, err := os.Stat(config.GetLibDirectory()); err != nil {
		t.Errorf("expected lib directory to exist, got %v", err)
	}
}

func TestAssertCacheDirectory(t *testing.T) {
	clearConfigEnv(t)

	config := NewDefaultConfig(WithPrefix(t.TempDir()))
	config.SetGlobalConfig()
	config.AssertCacheDirectory()

	if _, err := os.Stat(config.GetCacheDirectory()); err != nil {
		t.Errorf("expected cache directory to exist, got %v", err)
	}
}

func TestAssertMetadataFileExists(t *testing.T) {
	clearConfigEnv(t)

	config := NewDefaultConfig(WithPrefix(t.TempDir()))
	config.SetGlobalConfig()
	config.AssertCacheDirectory()

	metadataFile := config.GetMetadataFile()
	if err := os.WriteFile(metadataFile, []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to write metadata file: %v", err)
	}
	if err := os.WriteFile(metadataFile+".sigstore.json", []byte("{}"), 0o600); err != nil {
		t.Fatalf("failed to write metadata signature: %v", err)
	}

	config.AssertMetadataFileExists()
}
