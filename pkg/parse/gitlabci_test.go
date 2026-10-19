package parse

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

func TestParseServices(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected []GitLabCiService
	}{
		{
			name:     "empty services",
			input:    []any{},
			expected: []GitLabCiService{},
		},
		{
			name: "string services",
			input: []any{
				"docker:27-dind",
				"registry.example.com/custom:1.0.0",
			},
			expected: []GitLabCiService{
				{Name: "docker:27-dind"},
				{Name: "registry.example.com/custom:1.0.0"},
			},
		},
		{
			name: "map services",
			input: []any{
				map[string]any{"name": "docker:27-dind"},
				map[string]any{"name": "registry.example.com/custom:1.0.0"},
			},
			expected: []GitLabCiService{
				{Name: "docker:27-dind"},
				{Name: "registry.example.com/custom:1.0.0"},
			},
		},
		{
			name: "mixed services",
			input: []any{
				"docker:27-dind",
				map[string]any{"name": "registry.example.com/custom:1.0.0", "alias": "custom"},
			},
			expected: []GitLabCiService{
				{Name: "docker:27-dind"},
				{Name: "registry.example.com/custom:1.0.0"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseServices(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("expected %+v, got %+v", tt.expected, got)
			}
		})
	}
}

var gitlabCiPipelineYaml = `
include:
  - local: workflows/build.yml
workflow:
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
default:
  image: alpine:3.20
  services:
    - name: docker:27-dind
build:
  image: ghcr.io/uniget-org/tools/jq:1.7.1
  services:
    - docker:27-dind
    - name: registry.example.com/custom:1.0.0
      alias: custom
test:
  image: alpine:3.19
`

func TestLoadGitlabCi(t *testing.T) {
	pipeline, err := LoadGitlabCi([]byte(gitlabCiPipelineYaml))
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	if pipeline.Defaults.Image != "alpine:3.20" {
		t.Errorf("expected default image alpine:3.20, got %s", pipeline.Defaults.Image)
	}
	if len(pipeline.Defaults.Services) != 1 || pipeline.Defaults.Services[0].Name != "docker:27-dind" {
		t.Errorf("expected one default service docker:27-dind, got %+v", pipeline.Defaults.Services)
	}

	if len(pipeline.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(pipeline.Jobs))
	}

	build, ok := pipeline.Jobs["build"]
	if !ok {
		t.Fatalf("expected build job")
	}
	if build.Image != "ghcr.io/uniget-org/tools/jq:1.7.1" {
		t.Errorf("expected build image ghcr.io/uniget-org/tools/jq:1.7.1, got %s", build.Image)
	}
	if len(build.Services) != 2 {
		t.Fatalf("expected 2 build services, got %d", len(build.Services))
	}
	if build.Services[0].Name != "docker:27-dind" {
		t.Errorf("expected first build service docker:27-dind, got %s", build.Services[0].Name)
	}
	if build.Services[1].Name != "registry.example.com/custom:1.0.0" {
		t.Errorf("expected second build service registry.example.com/custom:1.0.0, got %s", build.Services[1].Name)
	}

	test, ok := pipeline.Jobs["test"]
	if !ok {
		t.Fatalf("expected test job")
	}
	if test.Image != "alpine:3.19" {
		t.Errorf("expected test image alpine:3.19, got %s", test.Image)
	}
}

func TestLoadGitlabCiWithInvalidYaml(t *testing.T) {
	_, err := LoadGitlabCi([]byte(":\n\t:sd dewed: ]]:]"))
	if err == nil {
		t.Fatalf("expected error for invalid yaml, got nil")
	}
}

func TestLoadGitlabCiWithWorkflowOnly(t *testing.T) {
	pipeline, err := LoadGitlabCi([]byte("workflow:\n  rules:\n    - if: \"true\""))
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(pipeline.Defaults.Services) != 0 || pipeline.Defaults.Image != "" {
		t.Errorf("expected empty defaults, got %+v", pipeline.Defaults)
	}
	if len(pipeline.Jobs) != 0 {
		t.Errorf("expected no jobs, got %+v", pipeline.Jobs)
	}
}

func TestLoadGitlabCiFromFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), ".gitlab-ci.yml")
	if err := os.WriteFile(filename, []byte(gitlabCiPipelineYaml), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	pipeline, err := LoadGitlabCiFromFile(filename)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if pipeline.Defaults.Image != "alpine:3.20" {
		t.Errorf("expected default image alpine:3.20, got %s", pipeline.Defaults.Image)
	}
}

func TestLoadGitlabCiFromMissingFile(t *testing.T) {
	_, err := LoadGitlabCiFromFile(filepath.Join(t.TempDir(), "does-not-exist.yml"))
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestExtractImageReferencesFromGitlabCi(t *testing.T) {
	pipeline, err := LoadGitlabCi([]byte(gitlabCiPipelineYaml))
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	expected := map[string]bool{}
	for _, image := range []string{
		"alpine:3.20",
		"docker:27-dind",
		"registry.example.com/custom:1.0.0",
		"ghcr.io/uniget-org/tools/jq:1.7.1",
		"alpine:3.19",
	} {
		imageRef, err := ref.New(image)
		if err != nil {
			t.Fatalf("failed to create ref from %q: %v", image, err)
		}
		expected[imageRef.Reference] = true
	}

	imageRefs, err := ExtractImageReferencesFromGitlabCi(pipeline)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(imageRefs.Refs) != len(expected) {
		t.Fatalf("expected %d refs, got %d: %+v", len(expected), len(imageRefs.Refs), imageRefs.Refs)
	}
	for _, imageRef := range imageRefs.Refs {
		if !expected[imageRef.Reference] {
			t.Errorf("unexpected ref %s", imageRef.Reference)
		}
	}
}

func TestExtractImageReferencesFromGitlabCiSkipsInvalidRefs(t *testing.T) {
	pipeline, err := LoadGitlabCi([]byte("build:\n  image: \"&&&invalid&&&\""))
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	imageRefs, err := ExtractImageReferencesFromGitlabCi(pipeline)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(imageRefs.Refs) != 0 {
		t.Errorf("expected no refs, got %+v", imageRefs.Refs)
	}
}

func TestBumpGitlabCiFileWithMissingFile(t *testing.T) {
	tools := &tool.Tools{}
	err := BumpGitlabCiFile(filepath.Join(t.TempDir(), "does-not-exist.yml"), tools, nil)
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestBumpGitlabCiFileWithoutImageReferences(t *testing.T) {
	filename := filepath.Join(t.TempDir(), ".gitlab-ci.yml")
	if err := os.WriteFile(filename, []byte("workflow:\n  rules:\n    - if: \"true\""), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	tools := &tool.Tools{}
	err := BumpGitlabCiFile(filename, tools, nil)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
}

func TestBumpGitlabCiFileWithMatchedTool(t *testing.T) {
	filename := filepath.Join(t.TempDir(), ".gitlab-ci.yml")
	if err := os.WriteFile(filename, []byte("build:\n  image: ghcr.io/uniget-org/tools/doesnotexist:1.0.0\n"), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	tools := &tool.Tools{}
	err := BumpGitlabCiFile(filename, tools, nil)
	if err == nil {
		t.Fatalf("expected error for unknown tool, got nil")
	}
	if !strings.Contains(err.Error(), "not found in metadata") {
		t.Errorf("expected not found error, got %s", err)
	}
}

func TestBumpGitlabCiFileWithUnigetImage(t *testing.T) {
	filename := filepath.Join(t.TempDir(), ".gitlab-ci.yml")
	if err := os.WriteFile(filename, []byte("build:\n  image: ghcr.io/uniget-org/tools/jq:1.7.1\n"), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	tools := &tool.Tools{
		Tools: []tool.Tool{
			{Name: "jq", Version: "1.7.1"},
		},
	}
	err := BumpGitlabCiFile(filename, tools, nil)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	// #nosec G304 -- reading test fixture that was just written
	newContent, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(newContent), "ghcr.io/uniget-org/tools/jq:1.7.1@sha256:") {
		t.Errorf("expected bumped reference with digest, got %s", string(newContent))
	}
}
