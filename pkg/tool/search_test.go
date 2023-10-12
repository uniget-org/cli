package tool

import (
	"strings"
	"testing"
)

var testSearchToolsString = `{
	"tools": [
		{
			"name":"foo",
			"version":"1.0.0",
			"runtime_dependencies": [
				"bar"
			],
			"tags": [
				"baz",
				"blarg"
			]
		},
		{
			"name":"bar",
			"version":"2.0.0",
			"tags": [
				"baz",
				"blubb"
			]
		}
	]
}`

func TestSearchContains(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testSearchToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	if !tools.Contains("foo") {
		t.Errorf("Tools should contain foo")
	}

	if !tools.Contains("bar") {
		t.Errorf("Tools should contain bar")
	}
}

func TestSearchGetByName(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testSearchToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	foo, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool foo: %s\n", err)
	}
	if foo.Name != "foo" {
		t.Errorf("Expected foo, got %s", foo.Name)
	}

	bar, err := tools.GetByName("bar")
	if err != nil {
		t.Errorf("Error getting tool bar: %s\n", err)
	}
	if bar.Name != "bar" {
		t.Errorf("Expected bar, got %s", bar.Name)
	}
}

func TestSearchGetByTag(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testSearchToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	toolList := tools.GetByTag("blarg")
	if len(toolList.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(toolList.Tools))
	}

	toolList = tools.GetByTag("baz")
	if len(toolList.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(toolList.Tools))
	}
}

func TestSearchGetByNames(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testSearchToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	toolList := tools.GetByNames([]string{"foo", "bar"})
	if len(toolList.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(toolList.Tools))
	}
}

func TestSearchGetByTags(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testSearchToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	toolList := tools.GetByTags([]string{"blarg", "blubb"})
	if len(toolList.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(toolList.Tools))
	}
}

func TestFind(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testSearchToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	toolList := tools.Find("foo", true, false, false)
	if len(toolList.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(toolList.Tools))
	}
	if toolList.Tools[0].Name != "foo" {
		t.Errorf("Expected tool foo, got %s", toolList.Tools[0].Name)
	}
	toolList.Tools[0].Print()

	toolList = tools.Find("baz", false, true, false)
	if len(toolList.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(toolList.Tools))
	}
	if toolList.Tools[0].Name != "foo" {
		t.Errorf("Expected tool foo, got %s", toolList.Tools[0].Name)
	}
	if toolList.Tools[1].Name != "bar" {
		t.Errorf("Expected tool bar, got %s", toolList.Tools[1].Name)
	}

	toolList = tools.Find("blubb", false, true, false)
	if len(toolList.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(toolList.Tools))
	}
	if toolList.Tools[0].Name != "bar" {
		t.Errorf("Expected tool bar, got %s", toolList.Tools[0].Name)
	}

	toolList = tools.Find("bar", false, false, true)
	toolList.List()
	if len(toolList.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(toolList.Tools))
	}
	if toolList.Tools[0].Name != "foo" {
		t.Errorf("Expected tool foo, got %s", toolList.Tools[0].Name)
	}
}

func TestGetNames(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testSearchToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	names := tools.GetNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}

	if names[0] != "foo" {
		t.Errorf("Expected foo, got %s", names[0])
	}

	if names[1] != "bar" {
		t.Errorf("Expected bar, got %s", names[1])
	}
}

func TestAddIfMissing(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testSearchToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	tools.AddIfMissing(&Tool{Name: "baz"})
	if len(tools.Tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools.Tools))
	}
}

func TestResolveDependencies(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testSearchToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	var plannedTools Tools
	err = tools.ResolveDependencies(&plannedTools, "foo")
	if err != nil {
		t.Errorf("Error resolving dependencies: %s\n", err)
	}
	if len(plannedTools.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d: %s", len(plannedTools.Tools), strings.Join(plannedTools.GetNames(), ","))
	}
}
