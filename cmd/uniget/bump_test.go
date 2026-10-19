package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bumpFixture is a file with only image references to tools which are not
// managed by uniget. Bumping these is a pass-through and does not require
// network access.
const bumpFixture = "docker:27-dind"

func TestBumpDockerfile(t *testing.T) {
	seedMetadata(t)

	dockerfile := filepath.Join(t.TempDir(), "Dockerfile")
	content := "FROM " + bumpFixture + "\n"
	if err := os.WriteFile(dockerfile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write dockerfile: %v", err)
	}

	if _, err := runCobraTest(t, nil, "bump", "dockerfile", "--file", dockerfile); err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}

	data, err := os.ReadFile(dockerfile) // #nosec G304 -- Reading test fixture
	if err != nil {
		t.Fatalf("failed to read dockerfile: %v", err)
	}
	if !strings.Contains(string(data), bumpFixture) {
		t.Errorf("expected dockerfile to still contain <%s>, received <%s>", bumpFixture, string(data))
	}
}

func TestBumpComposeFile(t *testing.T) {
	seedMetadata(t)

	composeFile := filepath.Join(t.TempDir(), "compose.yaml")
	content := "services:\n  app:\n    image: " + bumpFixture + "\n"
	if err := os.WriteFile(composeFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write compose file: %v", err)
	}

	if _, err := runCobraTest(t, nil, "bump", "compose", "--file", composeFile); err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}

	data, err := os.ReadFile(composeFile) // #nosec G304 -- Reading test fixture
	if err != nil {
		t.Fatalf("failed to read compose file: %v", err)
	}
	if !strings.Contains(string(data), bumpFixture) {
		t.Errorf("expected compose file to still contain <%s>, received <%s>", bumpFixture, string(data))
	}
}

func TestBumpKubernetesFile(t *testing.T) {
	seedMetadata(t)

	manifestFile := filepath.Join(t.TempDir(), "manifest.yaml")
	content := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n    - name: app\n      image: " + bumpFixture + "\n"
	if err := os.WriteFile(manifestFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	if _, err := runCobraTest(t, nil, "bump", "kubernetes", "--file", manifestFile); err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}

	data, err := os.ReadFile(manifestFile) // #nosec G304 -- Reading test fixture
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}
	if !strings.Contains(string(data), bumpFixture) {
		t.Errorf("expected manifest to still contain <%s>, received <%s>", bumpFixture, string(data))
	}
}

func TestBumpGitlabCiFile(t *testing.T) {
	seedMetadata(t)

	gitlabCiFile := filepath.Join(t.TempDir(), ".gitlab-ci.yml")
	content := "build:\n  stage: build\n  image: " + bumpFixture + "\n  script:\n    - echo hello\n"
	if err := os.WriteFile(gitlabCiFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write GitLab CI file: %v", err)
	}

	if _, err := runCobraTest(t, nil, "bump", "gitlab-ci", "--file", gitlabCiFile); err != nil {
		t.Fatalf("returned unexpected error: %v", err)
	}

	data, err := os.ReadFile(gitlabCiFile) // #nosec G304 -- Reading test fixture
	if err != nil {
		t.Fatalf("failed to read GitLab CI file: %v", err)
	}
	if !strings.Contains(string(data), bumpFixture) {
		t.Errorf("expected GitLab CI file to still contain <%s>, received <%s>", bumpFixture, string(data))
	}
}

func TestOutputCallback(t *testing.T) {
	// Output is printed via pterm to stdout. This is a smoke test.
	outputCallback("jq", "1.7.0", "1.7.1")
	outputCallback("baselayout", "3.5.0", "3.5.1")
}
