package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveUserCronTab(t *testing.T) {
	tt := []struct {
		name     string
		lines    []string
		expected []string
	}{
		{
			name:     "empty crontab",
			lines:    []string{},
			expected: []string{},
		},
		{
			name:     "removes uniget entries and reverses order",
			lines:    []string{"# m h dom mon dow command", "30 0 * * * uniget upgrade --auto-update", "keepme"},
			expected: []string{"keepme", "# m h dom mon dow command"},
		},
		{
			name:     "only uniget entries",
			lines:    []string{"30 0 * * * uniget upgrade --auto-update", "0 0 * * * uniget self-upgrade"},
			expected: []string{},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := removeUserCronTab(tc.lines)
			if len(actual) != len(tc.expected) {
				t.Fatalf("expected %v, received %v", tc.expected, actual)
			}
			for index := range tc.expected {
				if actual[index] != tc.expected[index] {
					t.Errorf("expected %v, received %v", tc.expected, actual)
					break
				}
			}
		})
	}
}

func TestCronCreateAndRemove(t *testing.T) {
	resetUnigetEnv(t)

	// Replace the crontab binary with a fake one to avoid manipulating the
	// crontab of the user running the tests.
	fakeBin := t.TempDir()
	script := `#!/bin/sh
if [ -n "$UNIGET_TEST_CRONTAB_FAIL" ]; then
	echo "simulated crontab failure" >&2
	exit 1
fi
if [ "$1" = "-l" ]; then
	printf 'keepme\n30 0 * * * uniget --user=false upgrade --auto-update\n'
fi
exit 0
`
	if err := os.WriteFile(filepath.Join(fakeBin, "crontab"), []byte(script), 0o700); err != nil { // #nosec G306 -- Test fixture needs executable bit
		t.Fatalf("failed to write fake crontab: %v", err)
	}
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	for _, args := range [][]string{
		{"cron", "create"},
		{"cron", "remove"},
	} {
		if _, err := runCobraTest(t, nil, args...); err != nil {
			t.Errorf("command %v returned unexpected error: %v", args, err)
		}
	}
}

func TestCronCreateFailure(t *testing.T) {
	resetUnigetEnv(t)

	fakeBin := t.TempDir()
	script := `#!/bin/sh
exit 1
`
	if err := os.WriteFile(filepath.Join(fakeBin, "crontab"), []byte(script), 0o700); err != nil { // #nosec G306 -- Test fixture needs executable bit
		t.Fatalf("failed to write fake crontab: %v", err)
	}
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	for _, args := range [][]string{
		{"cron", "create"},
		{"cron", "remove"},
	} {
		_, err := runCobraTest(t, nil, args...)
		if err == nil {
			t.Errorf("command %v did not return expected error", args)
		} else if !strings.Contains(err.Error(), "cannot get user crontab") {
			t.Errorf("unexpected error for %v, received <%s>", args, err.Error())
		}
	}
}
