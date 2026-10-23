package tool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newStatusTool(name string, version string, binary string, check string) *Tool {
	return &Tool{
		Name:    name,
		Version: version,
		Binary:  binary,
		Check:   check,
	}
}

func TestGetBinaryStatus(t *testing.T) {
	tmpDir := t.TempDir()

	binaryPath := filepath.Join(tmpDir, "binary")
	if err := os.WriteFile(binaryPath, []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to create binary file: %s", err)
	}

	blockerPath := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blockerPath, []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to create blocker file: %s", err)
	}

	tests := []struct {
		name          string
		binary        string
		expectPresent bool
		expectError   bool
	}{
		{
			name:          "no binary",
			binary:        "false",
			expectPresent: false,
		},
		{
			name:          "binary present",
			binary:        binaryPath,
			expectPresent: true,
		},
		{
			name:          "binary missing",
			binary:        filepath.Join(tmpDir, "nonexistent"),
			expectPresent: false,
		},
		{
			name:        "stat error",
			binary:      filepath.Join(blockerPath, "child"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := newStatusTool("tool", "1.0.0", tt.binary, "")
			err := tool.GetBinaryStatus()

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if tool.Status.BinaryPresent != tt.expectPresent {
				t.Errorf("expected BinaryPresent to be %v", tt.expectPresent)
			}
		})
	}
}

func TestGetMarkerFileStatus(t *testing.T) {
	tmpDir := t.TempDir()
	markerFileDirectory := filepath.Join(tmpDir, "markers")

	toolDir := filepath.Join(markerFileDirectory, "installed")
	if err := os.MkdirAll(toolDir, 0o750); err != nil {
		t.Fatalf("failed to create marker file directory: %s", err)
	}
	for _, version := range []string{"1.0.1", "1.0.2"} {
		if err := os.WriteFile(filepath.Join(toolDir, version), []byte(""), 0o600); err != nil {
			t.Fatalf("failed to create marker file: %s", err)
		}
	}

	blockerDir := filepath.Join(markerFileDirectory, "blocker")
	if err := os.MkdirAll(blockerDir, 0o750); err != nil {
		t.Fatalf("failed to create blocker directory: %s", err)
	}
	if err := os.WriteFile(filepath.Join(markerFileDirectory, "nondir"), []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create blocker file: %s", err)
	}

	tests := []struct {
		name              string
		version           string
		setupToolName     string
		expectError       bool
		expectMarkerSeen  bool
		expectMarkerVer   string
		expectVersionMach bool
	}{
		{
			name:             "no marker file",
			version:          "1.0.0",
			setupToolName:    "nothing",
			expectMarkerSeen: false,
		},
		{
			name:              "marker file with matching version",
			version:           "1.0.2",
			setupToolName:     "installed",
			expectMarkerSeen:  true,
			expectMarkerVer:   "1.0.2",
			expectVersionMach: true,
		},
		{
			name:              "marker file with different version",
			version:           "2.0.0",
			setupToolName:     "installed",
			expectMarkerSeen:  true,
			expectMarkerVer:   "1.0.2",
			expectVersionMach: false,
		},
		{
			name:          "marker directory is a file",
			version:       "1.0.0",
			setupToolName: "nondir",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := Tool{Name: tt.setupToolName, Version: tt.version}
			err := tool.GetMarkerFileStatus(markerFileDirectory)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if tool.Status.MarkerFilePresent != tt.expectMarkerSeen {
				t.Errorf("expected MarkerFilePresent to be %v", tt.expectMarkerSeen)
			}
			if tool.Status.MarkerFileVersion != tt.expectMarkerVer {
				t.Errorf("expected MarkerFileVersion to be %q, got %q", tt.expectMarkerVer, tool.Status.MarkerFileVersion)
			}
			if tool.Status.VersionMatches != tt.expectVersionMach {
				t.Errorf("expected VersionMatches to be %v", tt.expectVersionMach)
			}
		})
	}
}

func TestMarkerFilesCreateAndRemove(t *testing.T) {
	tmpDir := t.TempDir()
	markerFileDirectory := filepath.Join(tmpDir, "markers")

	blocker := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(blocker, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create blocker file: %s", err)
	}

	tool := newStatusTool("tool", "1.0.0", "", "")

	if err := tool.CreateMarkerFile(markerFileDirectory); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if _, err := os.Stat(filepath.Join(markerFileDirectory, "tool", "1.0.0")); err != nil {
		t.Errorf("expected marker file to exist: %v", err)
	}

	if err := tool.CreateMarkerFile(blocker); err == nil {
		t.Errorf("expected error when marker file directory is blocked, got nil")
	}

	if err := tool.RemoveMarkerFile(markerFileDirectory); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if _, err := os.Stat(filepath.Join(markerFileDirectory, "tool", "1.0.0")); !os.IsNotExist(err) {
		t.Errorf("expected marker file to be removed, got %v", err)
	}

	// Removing a non-existing marker file must not fail
	if err := tool.RemoveMarkerFile(markerFileDirectory); err != nil {
		t.Errorf("expected no error for missing marker file, got %s", err)
	}
}

func TestRunVersionCheck(t *testing.T) {
	tests := []struct {
		name          string
		check         string
		expectVersion string
		expectError   bool
	}{
		{
			name:          "successful version check",
			check:         "echo 1.2.3",
			expectVersion: "1.2.3",
		},
		{
			name:        "failing version check",
			check:       "((  ",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := newStatusTool("tool", "0.0.0", "", tt.check)
			version, err := tool.RunVersionCheck()

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if version != tt.expectVersion {
				t.Errorf("expected version %q, got %q", tt.expectVersion, version)
			}
		})
	}
}

func TestGetVersionStatus(t *testing.T) {
	tmpDir := t.TempDir()
	markerFileDirectory := filepath.Join(tmpDir, "markers")
	toolDir := filepath.Join(markerFileDirectory, "tool")
	if err := os.MkdirAll(toolDir, 0o750); err != nil {
		t.Fatalf("failed to create marker file directory: %s", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "1.0.0"), []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create marker file: %s", err)
	}

	binaryPath := filepath.Join(tmpDir, "binary")

	tests := []struct {
		name              string
		version           string
		check             string
		markerSeen        bool
		binaryPath        string
		expectVersion     string
		expectVersionMach bool
		expectError       bool
	}{
		{
			name:              "marker file preferred",
			version:           "1.0.0",
			check:             "echo 9.9.9",
			markerSeen:        true,
			expectVersion:     "1.0.0",
			expectVersionMach: true,
		},
		{
			name:              "version check with binary",
			version:           "1.2.3",
			check:             "echo 1.2.3",
			binaryPath:        binaryPath,
			expectVersion:     "1.2.3",
			expectVersionMach: true,
		},
		{
			name:              "version check failure",
			version:           "1.0.0",
			check:             "((  ",
			binaryPath:        binaryPath,
			expectVersionMach: false,
			expectError:       true,
		},
		{
			name:              "no marker and no check",
			version:           "1.0.0",
			markerSeen:        false,
			expectVersionMach: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := Tool{Name: tt.name, Version: tt.version, Check: tt.check}
			tool.Status.MarkerFilePresent = tt.markerSeen
			tool.Status.MarkerFileVersion = "1.0.0"
			if tt.binaryPath != "" {
				tool.Status.BinaryPresent = true
			}

			err := tool.GetVersionStatus()
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if tool.Status.Version != tt.expectVersion {
				t.Errorf("expected version %q, got %q", tt.expectVersion, tool.Status.Version)
			}
			if tool.Status.VersionMatches != tt.expectVersionMach {
				t.Errorf("expected VersionMatches to be %v", tt.expectVersionMach)
			}
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	tmpDir := t.TempDir()
	markerFileDirectory := filepath.Join(tmpDir, "markers")
	target := filepath.Join(tmpDir, "target")

	toolDir := filepath.Join(markerFileDirectory, "tool")
	if err := os.MkdirAll(toolDir, 0o750); err != nil {
		t.Fatalf("failed to create marker file directory: %s", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "1.0.0"), []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create marker file: %s", err)
	}

	binaryPath := filepath.Join(tmpDir, "binary")
	if err := os.WriteFile(binaryPath, []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to create binary file: %s", err)
	}

	tests := []struct {
		name              string
		version           string
		binary            string
		expectMarkerSeen  bool
		expectBinarySeen  bool
		expectVersion     string
		expectVersionMach bool
	}{
		{
			name:              "marker file based status",
			version:           "1.0.0",
			binary:            "false",
			expectBinarySeen:  false,
			expectMarkerSeen:  true,
			expectVersion:     "1.0.0",
			expectVersionMach: true,
		},
		{
			name:              "binary based status",
			version:           "9.9.9",
			binary:            binaryPath,
			expectBinarySeen:  true,
			expectMarkerSeen:  true,
			expectVersion:     "1.0.0",
			expectVersionMach: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := newStatusTool("tool", tt.version, tt.binary, "")
			err := tool.UpdateStatus(markerFileDirectory, target, markerFileDirectory, "x86_64", "x86_64")
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
			}
			if tool.Status.BinaryPresent != tt.expectBinarySeen {
				t.Errorf("expected BinaryPresent to be %v", tt.expectBinarySeen)
			}
			if tool.Status.MarkerFilePresent != tt.expectMarkerSeen {
				t.Errorf("expected MarkerFilePresent to be %v", tt.expectMarkerSeen)
			}
			if tool.Status.Version != tt.expectVersion {
				t.Errorf("expected version %q, got %q", tt.expectVersion, tool.Status.Version)
			}
			if tool.Status.VersionMatches != tt.expectVersionMach {
				t.Errorf("expected VersionMatches to be %v", tt.expectVersionMach)
			}
		})
	}

	// Error path: marker file status fails
	tool := newStatusTool("nondir", "1.0.0", "", "")
	if err := os.WriteFile(filepath.Join(markerFileDirectory, "nondir"), []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create blocker file: %s", err)
	}
	if err := tool.UpdateStatus(markerFileDirectory, target, markerFileDirectory, "x86_64", "x86_64"); err == nil {
		t.Errorf("expected error for bad marker file directory, got nil")
	}
}

func TestToolStateConvenience(t *testing.T) {
	tests := []struct {
		name             string
		binaryPresent    bool
		markerPresent    bool
		versionMatches   bool
		expectInstalled  bool
		expectUpgradable bool
		expectImportable bool
	}{
		{
			name:             "installed",
			binaryPresent:    true,
			markerPresent:    true,
			versionMatches:   true,
			expectInstalled:  true,
			expectUpgradable: false,
			expectImportable: false,
		},
		{
			name:             "upgradable",
			binaryPresent:    true,
			markerPresent:    true,
			versionMatches:   false,
			expectInstalled:  true,
			expectUpgradable: true,
			expectImportable: false,
		},
		{
			name:             "importable",
			binaryPresent:    true,
			markerPresent:    false,
			versionMatches:   false,
			expectInstalled:  false,
			expectUpgradable: false,
			expectImportable: true,
		},
		{
			name:             "not present",
			binaryPresent:    false,
			markerPresent:    false,
			versionMatches:   false,
			expectInstalled:  false,
			expectUpgradable: false,
			expectImportable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := Tool{}
			tool.Status = ToolStatus{
				BinaryPresent:     tt.binaryPresent,
				MarkerFilePresent: tt.markerPresent,
				VersionMatches:    tt.versionMatches,
			}

			if tool.IsInstalled() != tt.expectInstalled {
				t.Errorf("expected IsInstalled to be %v", tt.expectInstalled)
			}
			if tool.IsUpgradable() != tt.expectUpgradable {
				t.Errorf("expected IsUpgradable to be %v", tt.expectUpgradable)
			}
			if tool.IsImportable() != tt.expectImportable {
				t.Errorf("expected IsImportable to be %v", tt.expectImportable)
			}
		})
	}
}

func TestGetSources(t *testing.T) {
	tests := []struct {
		name               string
		sources            []Source
		fallbackRegistry   string
		fallbackRepository string
		expectRegistries   []string
		expectRepositories []string
	}{
		{
			name: "sources present",
			sources: []Source{
				{Registry: "ghcr.io", Repository: "foo/bar"},
				{Registry: "docker.io", Repository: "library/baz"},
			},
			expectRegistries:   []string{"ghcr.io", "docker.io"},
			expectRepositories: []string{"foo/bar", "library/baz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := Tool{Sources: tt.sources}

			registries, repositories := tool.GetSources()
			if strings.Join(registries, ",") != strings.Join(tt.expectRegistries, ",") {
				t.Errorf("expected registries %v, got %v", tt.expectRegistries, registries)
			}
			if strings.Join(repositories, ",") != strings.Join(tt.expectRepositories, ",") {
				t.Errorf("expected repositories %v, got %v", tt.expectRepositories, repositories)
			}

			registries, repositories = tool.GetSourcesWithFallback("fallback.io", "fallback/tool")
			expected := append(tt.expectRegistries, "fallback.io")
			if strings.Join(registries, ",") != strings.Join(expected, ",") {
				t.Errorf("expected registries %v, got %v", expected, registries)
			}
			expectedRepos := append(tt.expectRepositories, "fallback/tool")
			if strings.Join(repositories, ",") != strings.Join(expectedRepos, ",") {
				t.Errorf("expected repositories %v, got %v", expectedRepos, repositories)
			}
		})
	}
}

func TestReplaceVariables(t *testing.T) {
	tests := []struct {
		name      string
		binary    string
		target    string
		check     string
		expectBin string
		expectChk string
	}{
		{
			name:      "relative binary gets target prefix",
			binary:    "${target}/bin/${name}",
			target:    "/opt/target",
			check:     "${binary} --version of ${name}",
			expectBin: "/opt/target/bin/tool-name",
			expectChk: "/opt/target/bin/tool-name --version of tool-name",
		},
		{
			name:      "absolute binary keeps absolute path",
			binary:    "/bin/tool",
			target:    "/opt/target",
			check:     "${binary} --version",
			expectBin: "/bin/tool",
			expectChk: "/bin/tool --version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := Tool{Name: "tool-name", Binary: tt.binary, Check: tt.check}
			tool.ReplaceVariables(tt.target, "x86_64", "x86_64")

			if tool.Binary != tt.expectBin {
				t.Errorf("expected binary %q, got %q", tt.expectBin, tool.Binary)
			}
			if tool.Check != tt.expectChk {
				t.Errorf("expected check %q, got %q", tt.expectChk, tool.Check)
			}
		})
	}

	tool := Tool{Name: "tool", Binary: "false", Check: "irrelevant"}
	tool.ReplaceVariables("/opt/target", "x86_64", "x86_64")
	if tool.Binary != "false" {
		t.Errorf("expected binary \"false\" to stay untouched, got %q", tool.Binary)
	}

	// Empty binary must not panic and must stay empty
	emptyTool := Tool{Name: "tool"}
	emptyTool.ReplaceVariables("/opt/target", "x86_64", "x86_64")
	if emptyTool.Binary != "" {
		t.Errorf("expected empty binary to stay empty, got %q", emptyTool.Binary)
	}
}

func TestResolveDependenciesMissingAndSkipped(t *testing.T) {
	tests := []struct {
		name        string
		tools       []Tool
		target      string
		expectError bool
		expectQueue []string
	}{
		{
			name: "runtime dependency resolved first",
			tools: []Tool{
				{Name: "a", RuntimeDependencies: []string{"b"}},
				{Name: "b"},
			},
			target:      "a",
			expectQueue: []string{"b", "a"},
		},
		{
			name: "missing dependency is skipped",
			tools: []Tool{
				{Name: "a", RuntimeDependencies: []string{"missing"}},
			},
			target:      "a",
			expectQueue: []string{"a"},
		},
		{
			name:        "missing tool",
			tools:       []Tool{},
			target:      "unknown",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := Tools{Tools: tt.tools}
			queue := Tools{}

			err := tools.ResolveDependencies(&queue, tt.target)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %s", err)
			}

			if len(queue.Tools) != len(tt.expectQueue) {
				t.Fatalf("expected %d queued tools, got %d", len(tt.expectQueue), len(queue.Tools))
			}
			for index, name := range tt.expectQueue {
				if queue.Tools[index].Name != name {
					t.Errorf("expected queue entry %d to be %s, got %s", index, name, queue.Tools[index].Name)
				}
			}
		})
	}
}
