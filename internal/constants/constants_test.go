package constants

import (
	"strings"
	"testing"
)

func TestHeaderIsEmbedded(t *testing.T) {
	if Header == "" {
		t.Errorf("expected embedded header to be non-empty")
	}
}

func TestSlogan(t *testing.T) {
	if !strings.Contains(Slogan, "universal installer") {
		t.Errorf("expected slogan to contain 'universal installer'")
	}
	if !strings.Contains(Slogan, "https://uniget.dev") {
		t.Errorf("expected slogan to contain project URL")
	}
}

func TestProjectConstants(t *testing.T) {
	if ProjectName != "uniget" {
		t.Errorf("expected ProjectName to be uniget, got %s", ProjectName)
	}
	if MetadataFileName != "metadata.json" {
		t.Errorf("expected MetadataFileName to be metadata.json, got %s", MetadataFileName)
	}
	if MetadataImageTag != "main" {
		t.Errorf("expected MetadataImageTag to be main, got %s", MetadataImageTag)
	}
	if Registry != "ghcr.io" {
		t.Errorf("expected Registry to be ghcr.io, got %s", Registry)
	}
	if Organization != "uniget-org" {
		t.Errorf("expected Organization to be uniget-org, got %s", Organization)
	}
	if ImageRepository != "uniget-org/tools" {
		t.Errorf("expected ImageRepository to be uniget-org/tools, got %s", ImageRepository)
	}
	if ToolSeparator != "/" {
		t.Errorf("expected ToolSeparator to be /, got %s", ToolSeparator)
	}
	if RegistryImagePrefix != "ghcr.io/uniget-org/tools/" {
		t.Errorf("expected RegistryImagePrefix to be ghcr.io/uniget-org/tools/, got %s", RegistryImagePrefix)
	}
	if SigstoreIssuer != "https://token.actions.githubusercontent.com" {
		t.Errorf("expected SigstoreIssuer to be github token issuer, got %s", SigstoreIssuer)
	}
	if SigstoreSubjectRegexp == "" {
		t.Errorf("expected SigstoreSubjectRegexp to be non-empty")
	}
}

func TestHookDirectoryNames(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "pre-install", value: HooksPreInstallDirectory, expected: "hooks/pre-install.d"},
		{name: "post-install", value: HooksPostInstallDirectory, expected: "hooks/post-install.d"},
		{name: "pre-uninstall", value: HooksPreUninstallDirectory, expected: "hooks/pre-uninstall.d"},
		{name: "post-uninstall", value: HooksPostUninstallDirectory, expected: "hooks/post-uninstall.d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.value)
			}
		})
	}
}
