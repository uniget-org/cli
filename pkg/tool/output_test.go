package tool

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestListOne(t *testing.T) {
	var outBuffer bytes.Buffer

	tool := Tool{
		Name:    "test",
		Version: "1.2.3",
	}
	tool.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		"+---+------+---------+" + "\n" +
		"| # | NAME | VERSION |" + "\n" +
		"+---+------+---------+" + "\n" +
		"| 1 | test | 1.2.3   |" + "\n" +
		"+---+------+---------+" + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected %s, got %s", expectedOut, outBuffer.String())
	}
}

func TestListNone(t *testing.T) {
	var outBuffer bytes.Buffer

	tool := Tool{}
	tool.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		"+---+------+---------+" + "\n" +
		"| # | NAME | VERSION |" + "\n" +
		"+---+------+---------+" + "\n" +
		"| 1 |      |         |" + "\n" +
		"+---+------+---------+" + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected %s, got %s", expectedOut, outBuffer.String())
	}
}

func TestToolsListOne(t *testing.T) {
	var outBuffer bytes.Buffer

	tools := Tools{}
	tools.Tools = append(tools.Tools, Tool{
		Name:        "foo",
		Version:     "1.2.3",
		Description: "bar",
	})
	tools.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		" #  NAME  VERSION  DESCRIPTION " + "\n" +
		" 1  foo   1.2.3    bar         " + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected <%s>, got <%s>", expectedOut, outBuffer.String())
	}
}

func TestToolsListNone(t *testing.T) {
	var outBuffer bytes.Buffer

	tools := Tools{}
	tools.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		" #  NAME  VERSION  DESCRIPTION " + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected <%s>, got <%s>", expectedOut, outBuffer.String())
	}
}

func TestToolsListMultiple(t *testing.T) {
	var outBuffer bytes.Buffer

	tools := Tools{}
	tools.Tools = append(tools.Tools, Tool{
		Name:        "foo",
		Version:     "1.2.3",
		Description: "bar",
	})
	tools.Tools = append(tools.Tools, Tool{
		Name:        "baz",
		Version:     "1.2.3",
		Description: "blarg",
	})
	tools.List(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}

	expectedOut := "" +
		" #  NAME  VERSION  DESCRIPTION " + "\n" +
		" 1  foo   1.2.3    bar         " + "\n" +
		" 2  baz   1.2.3    blarg       " + "\n"

	if outBuffer.String() != expectedOut {
		t.Errorf("Expected <%s>, got <%s>", expectedOut, outBuffer.String())
	}
}

func TestToolsListWithStatus(t *testing.T) {
	var outBuffer bytes.Buffer

	tools := Tools{}
	tools.Tools = append(tools.Tools, Tool{
		Name:    "foo",
		Version: "1.2.3",
		Status: ToolStatus{
			BinaryPresent:     true,
			Version:           "1.2.3",
			VersionMatches:    true,
			MarkerFilePresent: true,
			MarkerFileVersion: "1.2.3",
			IsRequested:       true,
		},
	})
	tools.Tools = append(tools.Tools, Tool{
		Name:    "bar",
		Version: "2.0.0",
		Status: ToolStatus{
			IsRequested: true,
		},
	})
	tools.ListWithStatus(&outBuffer)

	if len(outBuffer.String()) == 0 {
		t.Errorf("outBuffer is empty")
	}
	if !strings.Contains(outBuffer.String(), "foo") || !strings.Contains(outBuffer.String(), "bar") {
		t.Errorf("expected tool names in output, got %q", outBuffer.String())
	}
	if !strings.Contains(outBuffer.String(), "1.2.3") || !strings.Contains(outBuffer.String(), "2.0.0") {
		t.Errorf("expected versions in output, got %q", outBuffer.String())
	}
}

func TestShowInternalsAndUsageAndUpdate(t *testing.T) {
	tool := Tool{
		Messages: Messages{
			Internals: "line1\nline2",
			Usage:     "usage text\n\nwith empty line\n",
			Update:    "update text",
		},
	}

	wantInternals := "      line1\n      line2\n"
	if got := tool.ShowInternals(6); got != wantInternals {
		t.Errorf("expected %q, got %q", wantInternals, got)
	}
	wantUsage := "      usage text\n      with empty line\n"
	if got := tool.ShowUsage(6); got != wantUsage {
		t.Errorf("expected %q, got %q", wantUsage, got)
	}
	wantUpdate := "  update text\n"
	if got := tool.ShowUpdate(2); got != wantUpdate {
		t.Errorf("expected %q, got %q", wantUpdate, got)
	}

	empty := Tool{}
	if got := empty.ShowInternals(6); got != "" {
		t.Errorf("expected empty internals, got %q", got)
	}
	if got := empty.ShowUsage(6); got != "" {
		t.Errorf("expected empty usage, got %q", got)
	}
	if got := empty.ShowUpdate(6); got != "" {
		t.Errorf("expected empty update, got %q", got)
	}
}

func TestPrintFullTool(t *testing.T) {
	var outBuffer bytes.Buffer

	tool := Tool{
		Name:        "jq",
		Version:     "1.7.1",
		Description: "CLI JSON processor",
		Homepage:    "https://jqlang.org",
		Repository:  "https://github.com/jqlang/jq",
		License:     License{Name: "MIT", Link: "https://opensource.org/license/mit"},
		Binary:      "false",
		Check:       "jq --version",
		Tags:        []string{"json", "cli"},
		BuildDependencies: []string{
			"gcc",
		},
		RuntimeDependencies: []string{},
		Platforms:           []string{"linux/amd64"},
		ConflictsWith:       []string{"gojq"},
		Messages: Messages{
			Internals: "internals text",
			Usage:     "usage text",
			Update:    "update text",
		},
		Renovate: Renovate{
			Datasource:     "github-tags",
			Package:        "jqlang/jq",
			ExtractVersion: "^jq-(?<version>.+)$",
			Versioning:     "loose",
		},
		Sources: []Source{
			{Registry: "ghcr.io", Repository: "uniget-org/tools"},
		},
		Lifecycle: Lifecycle{
			RenamedTo: "renamed-tool",
		},
		Status: ToolStatus{
			BinaryPresent:     true,
			Version:           "1.7.1",
			VersionMatches:    true,
			MarkerFilePresent: true,
			MarkerFileVersion: "1.7.1",
			IsRequested:       true,
		},
	}
	tool.Print(&outBuffer)

	out := outBuffer.String()
	for _, expected := range []string{
		"Name: jq\n",
		"  Description: CLI JSON processor\n",
		"  Homepage: https://jqlang.org\n",
		"  Repository: https://github.com/jqlang/jq\n",
		"  License: MIT (https://opensource.org/license/mit)",
		"  Version: 1.7.1\n",
		"  Binary: false\n",
		"  Check: <jq --version>\n",
		"    json\n",
		"  Build dependencies:\n    gcc\n",
		"  Conflicts with:\n    gojq\n",
		"  Platforms:\n    linux/amd64\n",
		"    Internals:\n",
		"    Usage:      usage text\n",
		"    Update:      update text\n",
		"  Renovate:\n",
		"    Datasource: github-tags\n",
		"    Package: jqlang/jq\n",
		"    ExtractVersion: ^jq-(?<version>.+)$\n",
		"    Versioning: loose\n",
		"  Sources:\n    ghcr.io/uniget-org/tools\n",
		"  Lifecycle:\n    Renamed to: renamed-tool\n",
		"    Binary present: true\n",
		"    Version: 1.7.1\n",
		"    Version matches: true\n",
		"    Marker file present: true\n",
		"    Marker file version: 1.7.1\n",
		"    Is requested: true\n",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q in output, missing", expected)
		}
	}
}

func TestPrintMinimalTool(t *testing.T) {
	var outBuffer bytes.Buffer

	tool := Tool{Name: "minimal"}
	tool.Print(&outBuffer)

	out := outBuffer.String()
	for _, expected := range []string{
		"Name: minimal\n",
		"  License: ",
		"  Tags:\n",
		"  Status\n",
		"    Binary present: false\n",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q in output, missing", expected)
		}
	}
	if strings.Contains(out, "Runtime dependencies:") ||
		strings.Contains(out, "Conflicts with:") ||
		strings.Contains(out, "Renovate:") ||
		strings.Contains(out, "Sources:") {
		t.Errorf("expected no optional sections, got %q", out)
	}
}

func TestPrintRemovedTool(t *testing.T) {
	var outBuffer bytes.Buffer

	tool := Tool{Name: "removed", Lifecycle: Lifecycle{RemovedWithReason: "upstream"}}
	tool.Print(&outBuffer)

	if !strings.Contains(outBuffer.String(), "Removed with reason: upstream") {
		t.Errorf("expected removed reason in output, got %q", outBuffer.String())
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestDescribe(t *testing.T) {
	tools := Tools{}
	tools.Tools = append(tools.Tools, Tool{Name: "jq"})

	var outBuffer bytes.Buffer
	if err := tools.Describe(&outBuffer, "jq"); err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	if !strings.Contains(outBuffer.String(), "Name:jq") {
		t.Errorf("expected tool information in output, got %q", outBuffer.String())
	}

	err := tools.Describe(&outBuffer, "nonexistent")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got %v", err)
	}

	if err := tools.Describe(failingWriter{}, "jq"); err == nil {
		t.Errorf("expected write error, got nil")
	}
}
