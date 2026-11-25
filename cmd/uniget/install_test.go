package main

import (
	"os"
	"path/filepath"
	"testing"

	myos "gitlab.com/uniget-org/cli/pkg/os"
)

func TestBinaryFileIsInUse(t *testing.T) {
	target := t.TempDir()
	binDir := filepath.Join(target, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil { // #nosec G301 -- Only test
		t.Fatalf("failed to create bin directory: %v", err)
	}
	binaryPath := filepath.Join(binDir, "tool")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil { // #nosec G306 -- Test fixture needs executable bit
		t.Fatalf("failed to create binary: %v", err)
	}

	if myos.BinaryFileIsInUse(target, binaryPath) {
		t.Errorf("expected writable binary to not be reported as in use")
	}

	// Paths outside of the target are never probed and never reported
	// as in use.
	outsidePath := filepath.Join(t.TempDir(), "elsewhere")
	if err := os.WriteFile(outsidePath, []byte("x"), 0o600); err != nil {
		t.Fatalf("failed to create file outside of target: %v", err)
	}
	if myos.BinaryFileIsInUse(target, outsidePath) {
		t.Errorf("expected binary outside of target to not be reported as in use")
	}

	// A missing binary inside the target cannot be opened for writing
	// and therefore behaves like a binary that is in use.
	if !myos.BinaryFileIsInUse(target, filepath.Join(binDir, "missing")) {
		t.Errorf("expected missing binary to be reported as in use")
	}

	// A directory cannot be opened for writing.
	if !myos.BinaryFileIsInUse(target, binDir) {
		t.Errorf("expected directory to be reported as in use")
	}
}
