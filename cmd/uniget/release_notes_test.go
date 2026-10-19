package main

import (
	"strings"
	"testing"
)

func TestExtractGitHubOwnerAndProject(t *testing.T) {
	tt := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "plain url",
			url:      "https://github.com/jqlang/jq",
			expected: "jqlang/jq",
		},
		{
			name:     "url with git suffix",
			url:      "https://github.com/jqlang/jq.git",
			expected: "jqlang/jq",
		},
		{
			name:     "git+https url",
			url:      "git+https://github.com/jqlang/jq.git",
			expected: "jqlang/jq",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := extractGitHubOwnerAndProject(tc.url)
			if err != nil {
				t.Fatalf("returned unexpected error: %v", err)
			}
			if actual != tc.expected {
				t.Errorf("expected %s, received %s", tc.expected, actual)
			}
		})
	}

	// URLs which do not match the pattern return an error.
	// Previously, they caused a panic.
	_, err := extractGitHubOwnerAndProject("https://gitlab.com/foo/bar")
	if err == nil {
		t.Fatalf("expected error for non-github url, got nil")
	}
	if !strings.Contains(err.Error(), "cannot extract owner and project") {
		t.Errorf("unexpected error, received <%s>", err.Error())
	}
}
