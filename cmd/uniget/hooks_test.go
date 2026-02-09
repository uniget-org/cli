package main

import (
	"fmt"
	"os"
	"testing"
)

func TestRunHook(t *testing.T) {
	dir := t.TempDir()
	hookFile := dir + "/test.sh"
	lockFile := dir + "/test.lck"

	content := fmt.Sprintf("#!/bin/bash\ntouch %s", lockFile)
	err := os.WriteFile(hookFile, []byte(content), 0755) // #nosec G306 -- Only test
	if err != nil {
		t.Fatalf("Failed to create hook file: %v", err)
	}

	_, err = runHook(hookFile)
	if err != nil {
		t.Fatalf("Failed to run hook: %s", err)
	}

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist, but it does not", lockFile)
	}
}

func TestProcessHooks(t *testing.T) {
	dir := t.TempDir()
	hookFile := dir + "/test.sh"
	lockFile := dir + "/test.lck"

	content := fmt.Sprintf("#!/bin/bash\ntouch %s", lockFile)
	err := os.WriteFile(hookFile, []byte(content), 0755) // #nosec G306 -- Only test
	if err != nil {
		t.Fatalf("Failed to create hook file: %v", err)
	}

	processHooks(dir, func(file string) error {
		_, err := runHook(file)
		return err
	})

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist, but it does not", lockFile)
	}
}
