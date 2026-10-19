package parse

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/tool"
)

const statefulSetManifest = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: example
spec:
  selector:
    matchLabels:
      app: example
  serviceName: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
      - name: example
        image: alpine:latest
`

func podManifest(toolImage string) string {
	return "apiVersion: v1\n" +
		"kind: Pod\n" +
		"metadata:\n" +
		"  name: example\n" +
		"spec:\n" +
		"  containers:\n" +
		"  - name: shell\n" +
		"    command: [\"sleep\", \"infinity\"]\n" +
		"    image: ghcr.io/uniget-org/tools/" + toolImage + ":1.0.0\n"
}

func TestBumpDockerfileWithMissingFile(t *testing.T) {
	tools := &tool.Tools{}
	err := BumpDockerfile(filepath.Join(t.TempDir(), "does-not-exist"), tools, nil)
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestBumpDockerfile(t *testing.T) {
	dockerfile := filepath.Join(t.TempDir(), "Dockerfile")
	content := "FROM docker:27-dind\n"
	if err := os.WriteFile(dockerfile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	tools := &tool.Tools{}
	err := BumpDockerfile(dockerfile, tools, nil)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	// #nosec G304 -- reading test fixture that was just written
	newContent, err := os.ReadFile(dockerfile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(newContent) != content {
		t.Errorf("expected file content to be unchanged, got %s", string(newContent))
	}
}

func TestBumpKubernetesFileWithMissingFile(t *testing.T) {
	tools := &tool.Tools{}
	err := BumpKubernetesFile(filepath.Join(t.TempDir(), "does-not-exist.yml"), tools, nil)
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestBumpKubernetesFileWithUnmatchedTool(t *testing.T) {
	manifest := filepath.Join(t.TempDir(), "pod.yml")
	if err := os.WriteFile(manifest, []byte(podManifest("doesnotexist")), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	tools := &tool.Tools{}
	err := BumpKubernetesFile(manifest, tools, nil)
	if err == nil {
		t.Fatalf("expected error for unknown tool, got nil")
	}
}

func TestLoadKubernetesManifestFromMissingFile(t *testing.T) {
	_, err := LoadKubernetesManifestFromFile(filepath.Join(t.TempDir(), "does-not-exist.yml"))
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestExtractImageReferencesFromKubernetesManifestWithUnsupportedKind(t *testing.T) {
	manifest, err := LoadKubernetesManifest([]byte(statefulSetManifest))
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	_, err = ExtractImageReferencesFromKubernetesManifest(manifest)
	if err == nil {
		t.Fatalf("expected error for unsupported manifest, got nil")
	}
}

func TestLoadComposeFileWithMissingFile(t *testing.T) {
	_, err := LoadComposeFile(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestBumpComposeFileWithMissingFile(t *testing.T) {
	tools := &tool.Tools{}
	err := BumpComposeFile(filepath.Join(t.TempDir(), "does-not-exist.yaml"), tools, nil)
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestBumpComposeFile(t *testing.T) {
	composeFile := filepath.Join(t.TempDir(), "compose.yaml")
	content := "services:\n" +
		"  dind:\n" +
		"    image: docker:27-dind\n"
	if err := os.WriteFile(composeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	tools := &tool.Tools{}
	err := BumpComposeFile(composeFile, tools, nil)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	// #nosec G304 -- reading test fixture that was just written
	newContent, err := os.ReadFile(composeFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(newContent) != content {
		t.Errorf("expected file content to be unchanged, got %s", string(newContent))
	}
}
