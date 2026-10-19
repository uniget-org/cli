package parse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

func TestReplaceInFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "dockerfile")
	content := []byte("FROM docker:27-dind\n")
	if err := os.WriteFile(filename, content, 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	reference, err := ref.New("docker:27-dind")
	if err != nil {
		t.Fatalf("failed to create ref: %v", err)
	}

	var callbackCalls int
	imageRefs := ImageRefs{}
	imageRefs.Add(reference)
	callback := func(toolName string, oldVersion string, newVersion string) {
		callbackCalls++
	}

	err = ReplaceInFile(filename, &imageRefs, &tool.Tools{}, callback)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	if callbackCalls != 0 {
		t.Errorf("expected no callback calls, got %d", callbackCalls)
	}

	// #nosec G304 -- reading test fixture that was just written
	newContent, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(newContent) != string(content) {
		t.Errorf("expected file content to be unchanged, got %s", string(newContent))
	}

	stat, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if stat.Mode().Perm() != 0o600 {
		t.Errorf("expected file mode to be preserved, got %v", stat.Mode().Perm())
	}
}

func TestReplaceInFileWithMissingFile(t *testing.T) {
	reference, err := ref.New("docker:27-dind")
	if err != nil {
		t.Fatalf("failed to create ref: %v", err)
	}

	imageRefs := ImageRefs{}
	imageRefs.Add(reference)

	err = ReplaceInFile(filepath.Join(t.TempDir(), "does-not-exist"), &imageRefs, &tool.Tools{}, nil)
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}
