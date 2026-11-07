package main

import (
	"fmt"
	"os"
	"testing"

	"gitlab.com/uniget-org/cli/internal/config"
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
