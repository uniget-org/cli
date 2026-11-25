package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/uniget-org/cli/internal/config"
	myos "gitlab.com/uniget-org/cli/pkg/os"
)

// setupHooksDirectory initializes the global configuration so that hooks must
// be placed inside GetHooksRootDirectory() as required by runHook. The
// returned directory is configuration.GetHooksRootDirectory().
func setupHooksDirectory(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	configuration = config.NewDefaultConfig(config.WithPrefix(dir))
	configuration.ConfigRoot = "etc"

	hooksDirectory := configuration.GetHooksRootDirectory()
	if err := os.MkdirAll(hooksDirectory, 0750); err != nil {
		t.Fatalf("Failed to create hooks directory %s: %v", hooksDirectory, err)
	}

	return hooksDirectory
}

func writeHookFile(t *testing.T, path string) {
	t.Helper()

	content := fmt.Sprintf("#!/bin/bash\ntouch %s", path+".lck")
	if err := os.WriteFile(path, []byte(content), 0755); err != nil { // #nosec G306 -- Only test
		t.Fatalf("Failed to create hook file: %v", err)
	}
}

func TestResolveHookFile(t *testing.T) {
	hooksDirectory := t.TempDir()

	tests := []struct {
		name    string
		hook    string
		wantErr bool
	}{
		{name: "plain file", hook: "test.sh"},
		{name: "nested file", hook: "sub/test.sh"},
		{name: "parent traversal", hook: "../escape.sh", wantErr: true},
		{name: "nested parent traversal", hook: "sub/../../escape.sh", wantErr: true},
		{name: "absolute path", hook: "/etc/passwd", wantErr: true},
		{name: "empty name", hook: "", wantErr: true},
		{name: "dot name", hook: ".", wantErr: true},
		{name: "sibling with shared prefix", hook: "../" + filepath.Base(hooksDirectory) + "-evil/test.sh", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := resolveHookFile(hooksDirectory, tt.hook)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got path %s", resolved)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if !myos.PathIsWithin(hooksDirectory, resolved) {
				t.Errorf("expected %s to be inside %s", resolved, hooksDirectory)
			}
		})
	}
}

func TestAddHooksCmdSanitizesSource(t *testing.T) {
	setupHooksDirectory(t)

	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "hook.sh")
	if err := os.WriteFile(sourceFile, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil { // #nosec G306 -- Test fixture needs executable bit
		t.Fatalf("failed to create hook source: %v", err)
	}
	sourceDir := filepath.Join(tempDir, "sourcedir")
	if err := os.Mkdir(sourceDir, 0o750); err != nil { // #nosec G301 -- Only test
		t.Fatalf("failed to create source directory: %v", err)
	}

	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{name: "regular file is accepted", source: sourceFile},
		{name: "missing file is rejected", source: filepath.Join(tempDir, "missing"), wantErr: true},
		{name: "directory is rejected", source: sourceDir, wantErr: true},
		{name: "trailing dot is rejected", source: tempDir + "/.", wantErr: true},
		{name: "trailing parent is rejected", source: tempDir + "/..", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hookType = "pre-install"
			hookSource = tt.source
			t.Cleanup(func() {
				hookType = ""
				hookSource = ""
			})

			err := addHooksCmd.RunE(addHooksCmd, []string{})
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if tt.wantErr {
				return
			}

			hookFile := filepath.Join(configuration.GetHooksPreInstallDirectory(), "hook.sh")
			info, statErr := os.Stat(hookFile)
			if statErr != nil {
				t.Fatalf("expected hook file %s to exist: %v", hookFile, statErr)
			}
			if info.Mode().Perm() != 0o700 {
				t.Errorf("expected mode 0700, got %s", info.Mode().Perm())
			}
		})
	}
}

func TestRunHook(t *testing.T) {
	hooksDirectory := setupHooksDirectory(t)
	hookFile := hooksDirectory + "/test.sh"
	lockFile := hookFile + ".lck"

	writeHookFile(t, hookFile)

	_, err := runHook(hookFile)
	if err != nil {
		t.Fatalf("Failed to run hook: %s", err)
	}

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist, but it does not", lockFile)
	}
}

func TestProcessHooks(t *testing.T) {
	hooksDirectory := setupHooksDirectory(t)
	hookFile := hooksDirectory + "/test.sh"
	lockFile := hookFile + ".lck"

	writeHookFile(t, hookFile)

	err := processHooks(hooksDirectory, func(file string) error {
		_, err := runHook(file)
		return err
	})
	if err != nil {
		t.Fatalf("Failed to process hooks: %s", err)
	}

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist, but it does not", lockFile)
	}
}
