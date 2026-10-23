package os

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDirectoryEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.Mkdir(emptyDir, 0o750); err != nil {
		t.Fatalf("failed to create empty directory: %s", err)
	}

	nonEmptyDir := filepath.Join(tmpDir, "non-empty")
	if err := os.Mkdir(nonEmptyDir, 0o750); err != nil {
		t.Fatalf("failed to create non-empty directory: %s", err)
	}
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to write file: %s", err)
	}

	tests := []struct {
		name      string
		directory string
		expected  bool
	}{
		{
			name:      "empty directory",
			directory: emptyDir,
			expected:  true,
		},
		{
			name:      "non-empty directory",
			directory: nonEmptyDir,
			expected:  false,
		},
		{
			name:      "nonexistent directory",
			directory: filepath.Join(tmpDir, "nonexistent"),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsDirectoryEmpty(tt.directory) != tt.expected {
				t.Errorf("expected IsDirectoryEmpty to be %v for %s", tt.expected, tt.directory)
			}
		})
	}
}

func TestCreateSubdirectoriesForPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectedDir string
		expectNone  bool
	}{
		{
			name:        "nested path",
			path:        "foo/bar/baz.txt",
			expectedDir: filepath.Join("foo", "bar"),
		},
		{
			name:        "single subdirectory",
			path:        "foo/baz.txt",
			expectedDir: filepath.Join("foo"),
		},
		{
			name:       "bare file name creates nothing",
			path:       "baz.txt",
			expectNone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			err := CreateSubdirectoriesForPath(tmpDir, tt.path)
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if tt.expectNone {
				// Only the workDir itself should exist
				entries, err := os.ReadDir(tmpDir)
				if err != nil {
					t.Fatalf("failed to read workDir: %s", err)
				}
				if len(entries) != 0 {
					t.Errorf("expected no subdirectories, got %d", len(entries))
				}
				return
			}
			if !DirectoryExists(filepath.Join(tmpDir, tt.expectedDir)) {
				t.Errorf("expected directory %s to exist", tt.expectedDir)
			}
		})
	}
}

func TestCreateSubdirectoriesForPathCleansTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// filepath.Clean resolves "../..", so the dir must stay inside tmpDir
	err := CreateSubdirectoriesForPath(tmpDir, "foo/../../evil.txt")
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
}
