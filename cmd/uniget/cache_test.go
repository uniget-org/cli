package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/uniget-org/cli/internal/config"
)

func TestCacheIsConfigured(t *testing.T) {
	savedConfiguration := configuration
	t.Cleanup(func() { configuration = savedConfiguration })
	configuration = config.NewDefaultConfig()

	tt := []struct {
		name     string
		cache    string
		expected bool
	}{
		{"none", "none", false},
		{"empty", "", false},
		{"file", "file", true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			configuration.Cache = tc.cache
			if actual := cacheIsConfigured(); actual != tc.expected {
				t.Errorf("expected %v, received %v", tc.expected, actual)
			}
		})
	}
}

func TestDirSize(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "file1"), make([]byte, 10), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(directory, "subdirectory"), 0o750); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "subdirectory", "file2"), make([]byte, 20), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	size, err := dirSize(directory)
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if size != 30 {
		t.Errorf("expected 30, received %d", size)
	}

	if _, err := dirSize(filepath.Join(directory, "doesnotexist")); err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestDirList(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "file1"), nil, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(directory, "subdirectory"), 0o750); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "subdirectory", "file2"), nil, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	files, err := dirList(directory)
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, received %d: %v", len(files), files)
	}
	for _, expected := range []string{"file1", "file2"} {
		found := false
		for _, file := range files {
			if file == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected files to contain <%s>, received %v", expected, files)
		}
	}

	if _, err := dirList(filepath.Join(directory, "doesnotexist")); err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestDirPrune(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "file1"), nil, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(directory, "subdirectory"), 0o750); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "subdirectory", "file2"), nil, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	count, err := dirPrune(directory)
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2, received %d", count)
	}

	files, err := dirList(directory)
	if err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected directory to be empty, received %v", files)
	}

	if _, err := dirPrune(filepath.Join(directory, "doesnotexist")); err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestCacheNotConfigured(t *testing.T) {
	resetUnigetEnv(t)
	_, err := runCobraTest(t, nil, "cache", "info")
	if err == nil {
		t.Error("expected error for unconfigured cache")
	} else if !strings.Contains(err.Error(), "cache is not configured") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}
}

func TestCacheWithFileCache(t *testing.T) {
	resetUnigetEnv(t)

	// Keep the working directory separate from the prefix so that a
	// CWD-relative cache directory can be detected.
	prefix := t.TempDir()
	workDirectory := t.TempDir()
	t.Chdir(workDirectory)

	t.Setenv("UNIGET_PREFIX", prefix)
	t.Setenv("UNIGET_CACHE", "file")
	t.Setenv("UNIGET_CACHEDIRECTORY", "downloads")

	downloads := filepath.Join(prefix, "downloads")

	// AssertDirectory in the persistent pre-run must create the cache
	// directory under the prefix. Previously, it was created relative to
	// the current working directory.
	if _, err := runCobraTest(t, nil, "cache", "info"); err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}
	if info, err := os.Stat(downloads); err != nil || !info.IsDir() {
		t.Errorf("expected cache directory %s to have been created: %v", downloads, err)
	}
	if _, err := os.Stat(filepath.Join(workDirectory, "downloads")); !os.IsNotExist(err) {
		t.Error("expected no cache directory in the working directory")
	}

	if err := os.MkdirAll(downloads, 0o750); err != nil {
		t.Fatalf("failed to create cache directory: %v", err)
	}
	if err := os.MkdirAll(downloads, 0o750); err != nil {
		t.Fatalf("failed to create cache directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(downloads, "blob1"), make([]byte, 10), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(downloads, "blob2"), make([]byte, 20), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Information is printed to stdout by these commands. Only the error
	// path is asserted here.
	for _, args := range [][]string{
		{"cache", "info"},
		{"cache", "stats"},
		{"cache", "list"},
		{"cache", "prune"},
	} {
		if _, err := runCobraTest(t, nil, args...); err != nil {
			t.Errorf("command %v returned unexpected error: %v", args, err)
		}
	}

	files, err := os.ReadDir(downloads)
	if err != nil {
		t.Fatalf("failed to read cache directory: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected cache directory to be empty, received %d entries", len(files))
	}
}

func TestCacheContainerdNotImplemented(t *testing.T) {
	resetUnigetEnv(t)
	t.Setenv("UNIGET_PREFIX", t.TempDir())
	t.Setenv("UNIGET_CACHE", "containerd")

	for _, args := range [][]string{
		{"cache", "stats"},
		{"cache", "list"},
		{"cache", "prune"},
	} {
		_, err := runCobraTest(t, nil, args...)
		if err == nil {
			t.Errorf("command %v did not return expected error", args)
		} else if !strings.Contains(err.Error(), "not yet implemented") {
			t.Errorf("unexpected error for %v, received <%s>", args, err.Error())
		}
	}
}
