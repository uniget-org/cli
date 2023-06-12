package tool

import (
	"testing"
)

var testLoadToolsString = `{
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
			],
			"homepage": "https://foo.bar",
			"description": "Foo"
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

func TestLoadFromBytes(t *testing.T) {
	tools, err := LoadFromBytes([]byte(testLoadToolsString))
	if err != nil {
		t.Errorf("Error loading data: %s\n", err)
	}

	if len(tools.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools.Tools))
	}

	foo, err := tools.GetByName("foo")
	if err != nil {
		t.Errorf("Error getting tool foo: %s\n", err)
	}

	foo.Print()

	if foo.Name != "foo" {
		t.Errorf("Expected foo, got %s", foo.Name)
	}
	if foo.Version != "1.0.0" {
		t.Errorf("Expected 1.0.0, got %s", foo.Version)
	}
	if len(foo.RuntimeDependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(foo.RuntimeDependencies))
	}
	if foo.RuntimeDependencies[0] != "bar" {
		t.Errorf("Expected bar, got %s", foo.RuntimeDependencies[0])
	}
}
