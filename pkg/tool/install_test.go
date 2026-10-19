package tool

import (
	"bytes"

	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/safearchive/tar"
)

func TestApplyPathRewrites(t *testing.T) {
	rules := []PathRewrite{
		{Source: "usr/local/", Target: "", Operation: "REPLACE"},
		{Source: "var/lib/uniget/", Target: "/var/lib/uniget/", Operation: "REPLACE", Abort: true},
		{Source: "bin/", Target: "opt/bin/", Operation: "REPLACE", Abort: true},
		{Source: "", Target: "target/", Operation: "PREPEND"},
		{Source: "unsupported/", Target: "x/", Operation: "INVALID"},
	}
	tests := []struct {
		name     string
		input    string
		rules    []PathRewrite
		expected string
	}{
		{
			name:     "replace at start",
			input:    "usr/local/bin/jq",
			expected: "opt/bin/jq",
		},
		{
			name:     "no rule applies, prepend",
			input:    "etc/profile",
			expected: "target/etc/profile",
		},
		{
			name:     "already prepended target",
			input:    "target/bin/jq",
			expected: "target/bin/jq",
		},
		{
			name:     "abort rule stops processing",
			input:    "var/lib/uniget/jq/metadata.json",
			expected: "/var/lib/uniget/jq/metadata.json",
		},
		{
			name:     "unsupported operation is skipped",
			input:    "unsupported/file",
			expected: "target/unsupported/file",
		},
		{
			name:     "replace with empty target on absolute path result",
			input:    "usr/local/etc/config",
			expected: "target/etc/config",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "target/",
		},
		{
			name:     "no rules",
			input:    "usr/local/bin/jq",
			rules:    []PathRewrite{},
			expected: "usr/local/bin/jq",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rulesForTest := rules
			if tt.rules != nil {
				rulesForTest = tt.rules
			}
			got := applyPathRewrites(tt.input, rulesForTest)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func buildTestLayer(t *testing.T) io.ReadCloser {
	t.Helper()

	buffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(buffer)

	entries := []struct {
		header  *tar.Header
		content string
	}{
		{header: &tar.Header{Name: "usr/local/", Typeflag: tar.TypeDir, Mode: 0o755}},
		{header: &tar.Header{Name: "usr/local/bin/jq", Typeflag: tar.TypeReg, Mode: 0o755}, content: "binary"},
		{header: &tar.Header{Name: "usr/local/bin/jq-link", Typeflag: tar.TypeSymlink, Linkname: "jq", Mode: 0o777}},
		{header: &tar.Header{Name: "usr/local/bin/jqx", Typeflag: tar.TypeLink, Linkname: "usr/local/bin/jq", Mode: 0o755}},
		{header: &tar.Header{Name: "usr/local/dev/zero", Typeflag: tar.TypeChar, Mode: 0o666}},
	}
	for _, entry := range entries {
		if entry.header.Size == 0 {
			entry.header.Size = int64(len(entry.content))
		}
		if err := tarWriter.WriteHeader(entry.header); err != nil {
			t.Fatalf("failed to write header: %v", err)
		}
		if entry.content != "" {
			if _, err := tarWriter.Write([]byte(entry.content)); err != nil {
				t.Fatalf("failed to write content: %v", err)
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	return io.NopCloser(bytes.NewReader(buffer.Bytes()))
}

func TestToolInspect(t *testing.T) {
	layer := buildTestLayer(t)
	rules := []PathRewrite{
		{Source: "usr/local/", Target: "", Operation: "REPLACE"},
	}

	result, err := toolFixture().Inspect(io.Discard, layer, rules)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	expected := []string{
		"-rwxr-xr-x bin/jq",
		"lrwxrwxrwx bin/jq-link -> jq",
		"lrwxr-xr-x bin/jqx -> usr/local/bin/jq",
		"Unknown: dev/zero",
	}
	if len(result) != len(expected) {
		t.Fatalf("expected %d entries, got %d: %v", len(expected), len(result), result)
	}
	for i, entry := range expected {
		if result[i] != entry {
			t.Errorf("entry %d: expected %q, got %q", i, entry, result[i])
		}
	}
}

func TestToolInspectWithCorruptTar(t *testing.T) {
	_, err := toolFixture().Inspect(io.Discard, io.NopCloser(strings.NewReader("not a gzip or tar")), nil)
	if err == nil {
		t.Fatalf("expected error for corrupt layer, got nil")
	}
}

func TestToolInstall(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	layer := buildTestLayer(t)
	rules := []PathRewrite{
		{Source: "usr/local/", Target: "", Operation: "REPLACE"},
	}
	patchFile := func(path string) string {
		return strings.Replace(path, "bin/", "binaries/", 1)
	}

	result, err := toolFixture().Install(io.Discard, layer, rules, patchFile)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	expected := []string{
		"binaries/jq",
		"binaries/jq-link",
		"binaries/jqx",
		"dev/zero",
	}
	if len(result) != len(expected) {
		t.Fatalf("expected %d installed files, got %d: %v", len(expected), len(result), result)
	}
	for i, entry := range expected {
		if result[i] != entry {
			t.Errorf("entry %d: expected %q, got %q", i, entry, result[i])
		}
	}

	// The regular file and the symlink must exist on disk after install.
	for _, file := range []string{"bin/jq", "bin/jq-link"} {
		if _, err := os.Lstat(file); err != nil {
			t.Errorf("expected %s to be extracted, got %v", file, err)
		}
	}
}

func TestToolInstallWithCorruptTar(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	_, err := toolFixture().Install(io.Discard, io.NopCloser(strings.NewReader("not a gzip or tar")), nil, func(path string) string { return path })
	if err == nil {
		t.Fatalf("expected error for corrupt layer, got nil")
	}
}

func toolFixture() *Tool {
	return &Tool{Name: "jq"}
}
