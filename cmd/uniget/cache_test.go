package main

import (
	"strings"
	"testing"
)

func TestCacheCommandsAreNoOps(t *testing.T) {
	resetUnigetEnv(t)
	t.Setenv("UNIGET_PREFIX", t.TempDir())

	for _, args := range [][]string{
		{"cache", "info"},
		{"cache", "stats"},
		{"cache", "list"},
		{"cache", "prune"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			out, err := runCobraTest(t, nil, args...)
			if err != nil {
				t.Fatalf("command %v returned unexpected error: %v", args, err)
			}
			if strings.TrimSpace(out) != "" {
				t.Errorf("command %v produced unexpected output: %q", args, out)
			}
		})
	}
}

func TestCacheCommandsAreNoOpsWithBackend(t *testing.T) {
	resetUnigetEnv(t)
	t.Setenv("UNIGET_PREFIX", t.TempDir())
	t.Setenv("UNIGET_CACHE", "file")
	t.Setenv("UNIGET_CACHEDIRECTORY", "downloads")

	for _, args := range [][]string{
		{"cache", "info"},
		{"cache", "stats"},
		{"cache", "list"},
		{"cache", "prune"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			out, err := runCobraTest(t, nil, args...)
			if err != nil {
				t.Fatalf("command %v returned unexpected error: %v", args, err)
			}
			if !strings.Contains(out, "Cache backend file is no longer supported") {
				t.Errorf("command %v did not report the deprecated backend: %q", args, out)
			}
		})
	}
}
