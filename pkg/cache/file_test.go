package cache

import (
	"os"
	"testing"
)

func TestNewFileCache(t *testing.T) {
	cache := NewFileCache("test")
	if cache.cacheDirectory != "test" {
		t.Errorf("expected cache directory to be 'test', got '%s'", cache.cacheDirectory)
	}
}

func TestCheckIfCacheDirectoryExists(t *testing.T) {
	dir := t.TempDir()
	_, err := os.Stat(dir)
	if err != nil {
		t.Errorf("go testing failed to provide temporary directory %s", dir)
	}
	t.Logf("temporary directory: %s", dir)
	t.Logf("temporary directory exists: %t", !os.IsNotExist(err))

	cache := NewFileCache(dir)
	if cache.cacheDirectory != dir {
		t.Errorf("expected cache directory to be '%s', got '%s'", dir, cache.cacheDirectory)
	}
	if !cache.cacheDirectoryExists() {
		t.Errorf("temporary cache directory does to exist %s", cache.cacheDirectory)
	}
}
