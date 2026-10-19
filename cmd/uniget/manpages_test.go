package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManpagesRelativePath(t *testing.T) {
	resetUnigetEnv(t)
	prefix := t.TempDir()
	t.Setenv("UNIGET_PREFIX", prefix)

	// Relative paths resolve against the target directory. Information
	// about progress is printed to stdout by the command.
	if _, err := runCobraTest(t, nil, "manpages"); err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}

	manDirectory := filepath.Join(prefix, "usr", "local", "share", "man", "man1")
	for _, expected := range []string{"uniget.1", "uniget-tags.1", "uniget-search.1"} {
		if _, err := os.Stat(filepath.Join(manDirectory, expected)); err != nil {
			t.Errorf("expected manpage %s to exist: %v", expected, err)
		}
	}
}

func TestManpagesAbsolutePath(t *testing.T) {
	resetUnigetEnv(t)
	manPath := filepath.Join(t.TempDir(), "man")
	if _, err := runCobraTest(t, nil, "manpages", "--path", manPath); err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(manPath, "man1", "uniget.1")); err != nil {
		t.Errorf("expected manpage uniget.1 to exist: %v", err)
	}
}
