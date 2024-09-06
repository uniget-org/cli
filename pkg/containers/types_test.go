package containers

import "testing"

func TestNewToolRef(t *testing.T) {
	ref := NewToolRef("a", "b", "c", "d")

	if ref.Registry != "a" {
		t.Errorf("Registry is invalid")
	}
	if ref.Repository != "b" {
		t.Errorf("Repository is invalid")
	}
	if ref.Tool != "c" {
		t.Errorf("Tool is invalid")
	}
	if ref.Version != "d" {
		t.Errorf("Version is invalid")
	}
}

func TestNewToolRefToString(t *testing.T) {
	ref := NewToolRef("a", "b", "c", "d")

	if ref.String() != "a/b/c:d" {
		t.Errorf("String is invalid")
	}
}

func TestNewToolRefKey(t *testing.T) {
	ref := NewToolRef(
		"a",
		"b",
		"c",
		"d",
	)
	if ref.Key() != "c-d" {
		t.Errorf("expected key to be 'c-d', got '%s'", ref.Key())
	}
}
