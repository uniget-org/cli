package tool

import (
	"testing"
)

var testToolToolsString = `{
	"tools": [
		{
			"name":    "foo",
			"version": "1.0.0",
			"binary":  "${target}/bin/foo",
			"check":   "${binary} --version",
			"build_dependencies": [
				"bar"
			],
			"runtime_dependencies": [
				"bar"
			],
			"tags": [
				"baz",
				"blarg"
			]
		},
		{
			"name":    "bar",
			"version": "2.0.0",
			"binary":  "baz",
			"tags": [
				"baz",
				"blubb"
			]
		},
		{
			"name": "myname1-suffix"
		},
		{
			"name": "myname2_suffix"
		},
		{
			"name": "myname3-suffix1-suffix2"
		},
		{
			"name": "myname4_suffix1-suffix2"
		}
	]
}`

func TestTool_CamelCaseName(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testToolToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tool, err := tools.GetByName("myname1-suffix")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}
	if tool.GetCamelCaseName() != "Myname1Suffix" {
		t.Errorf("Expected 'Myname1Suffix', got '%s'", tool.GetCamelCaseName())
	}

	tool, err = tools.GetByName("myname2_suffix")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}
	if tool.GetCamelCaseName() != "Myname2Suffix" {
		t.Errorf("Expected 'Myname2Suffix', got '%s'", tool.GetCamelCaseName())
	}

	tool, err = tools.GetByName("myname3-suffix1-suffix2")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}
	if tool.GetCamelCaseName() != "Myname3Suffix1Suffix2" {
		t.Errorf("Expected 'Myname3Suffix1Suffix2', got '%s'", tool.GetCamelCaseName())
	}

	tool, err = tools.GetByName("myname4_suffix1-suffix2")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}
	if tool.GetCamelCaseName() != "Myname4Suffix1Suffix2" {
		t.Errorf("Expected 'Myname4Suffix1Suffix2', got '%s'", tool.GetCamelCaseName())
	}
}

func TestTool_MatchesName(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testToolToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tool, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}

	if !tool.MatchesName("foo") {
		t.Errorf("Tools should contain foo")
	}
	if !tool.MatchesName("fo") {
		t.Errorf("Tools should contain fo")
	}
}

func TestTool_HasTag(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testToolToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tool, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}

	if !tool.HasTag("baz") {
		t.Errorf("Tools should contain baz")
	}
	if !tool.HasTag("blarg") {
		t.Errorf("Tools should contain blarg")
	}
}

func TestTool_MatchesTag(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testToolToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tool, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}

	if !tool.MatchesTag("baz") {
		t.Errorf("Tools should contain baz")
	}
	if !tool.MatchesTag("blarg") {
		t.Errorf("Tools should contain blarg")
	}
	if !tool.MatchesTag("ba") {
		t.Errorf("Tools should contain ba")
	}
}

func TestTool_HasBuildDependency(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testToolToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tool, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}

	if !tool.HasBuildDependency("bar") {
		t.Errorf("Tools should contain bar")
	}
}

func TestTool_HasRuntimeDependency(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testToolToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tool, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}

	if !tool.HasRuntimeDependency("bar") {
		t.Errorf("Tools should contain bar")
	}
}

func TestTool_MatchesBuildDependency(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testToolToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tool, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}

	if !tool.MatchesBuildDependency("bar") {
		t.Errorf("Tools should contain bar")
	}
	if !tool.MatchesBuildDependency("ba") {
		t.Errorf("Tools should contain ba")
	}
}

func TestTool_MatchesRuntimeDependency(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testToolToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tool, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}

	if !tool.MatchesRuntimeDependency("bar") {
		t.Errorf("Tools should contain bar")
	}
	if !tool.MatchesRuntimeDependency("ba") {
		t.Errorf("Tools should contain ba")
	}
}

func TestTool_replaceVariables(t *testing.T) {
	result := replaceVariables("foobarblargblubb", []string{"foo", "bar"}, []string{"blarg", "blubb"})
	if result != "blargblubbblargblubb" {
		t.Errorf("Expected 'blargblubbblargblubb', got '%s'", result)
	}
}

func TestTool_ReplaceVariables(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testToolToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tool, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}
	tool.ReplaceVariables("/usr/local", "x86_64", "amd64")
	if tool.Binary != "/usr/local/bin/foo" {
		t.Errorf("Expected '/usr/local/bin/foo', got '%s'", tool.Binary)
	}
	if tool.Check != "/usr/local/bin/foo --version" {
		t.Errorf("Expected '/usr/local/bin/foo --version', got '%s'", tool.Check)
	}

	tool, err = tools.GetByName("bar")
	if err != nil {
		t.Errorf("Error getting tool: %s\n", err)
	}
	tool.ReplaceVariables("/usr/local", "x86_64", "amd64")
	if tool.Binary != "/usr/local/bin/baz" {
		t.Errorf("Expected '/usr/local/bin/baz', got '%s'", tool.Binary)
	}
}
