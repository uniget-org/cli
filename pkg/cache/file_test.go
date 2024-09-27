package cache

import (
	"os"
	"testing"

	"github.com/uniget-org/cli/pkg/containers"
)

func TestToolRefKey(t *testing.T) {
	ref := containers.NewToolRef(
		"a",
		"b",
		"c",
		"d",
	)
	if ref.Key() != "c-d" {
		t.Errorf("expected key to be 'c-d', got '%s'", ref.Key())
	}
}

func TestNewFileCache(t *testing.T) {
	cache := NewFileCache("test", 300)
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

	cache := NewFileCache(dir, 300)
	if !cache.cacheDirectoryExists() {
		t.Errorf("temporary cache directory does not exist %s", cache.cacheDirectory)
	}
}

func TestFileCacheGetManually(t *testing.T) {
	cache := NewFileCache(t.TempDir(), 300)

	if cache.checkDataInCache(toolRef.Key()) {
		t.Errorf("unexpected cache hit")
	}
	_, err := cache.readDataFromCache(toolRef.Key())
	if err == nil {
		t.Errorf("cache should be empty")
	}

	var testData = []byte("test")

	err = cache.writeDataToCache(testData, toolRef.Key())
	if err != nil {
		t.Errorf("failed to write data to cache: %v", err)
	}
	if !cache.checkDataInCache(toolRef.Key()) {
		t.Errorf("cache miss")
	}
	data, err := cache.readDataFromCache(toolRef.Key())
	if err != nil {
		t.Errorf("failed to read key %s after cache hit: %v", toolRef.Key(), err)
	}
	if string(data) != string(testData) {
		t.Errorf("expected data to be 'test', got '%s'", string(data))
	}
}

func TestFileCacheGet(t *testing.T) {
	cache := NewFileCache(t.TempDir(), 300)

	if cache.checkDataInCache(toolRef.Key()) {
		t.Errorf("unexpected cache hit")
	}

	data, err := cache.Get(toolRef)
	if err != nil {
		t.Errorf("failed to get data from cache: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("expected data to be non-empty (first attempt)")
	}

	if !cache.checkDataInCache(toolRef.Key()) {
		t.Errorf("unexpected cache miss")
	}

	data, err = cache.Get(toolRef)
	if err != nil {
		t.Errorf("failed to get data from cache: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("expected data to be non-empty (second attempt)")
	}
}
