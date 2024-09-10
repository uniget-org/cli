package containers

import (
	"fmt"
	"testing"
)

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

func TestGetRef(t *testing.T) {
	registry := "foo.com"
	imageRepository := "b"
	tool := "c"
	version := "d"

	toolRef := NewToolRef(registry, imageRepository, tool, version)
	ref := toolRef.GetRef()
	t.Logf("ref: %v", ref)

	if ref.Registry != registry {
		t.Errorf("Registry is invalid, expected %s, got %s", registry, ref.Registry)
	}

	expectedImageRepository := fmt.Sprintf("%s/%s", imageRepository, tool)
	if ref.Repository != expectedImageRepository {
		t.Errorf("Repository is invalid, expected %s, got %s", imageRepository, ref.Repository)
	}

	if ref.Tag != version {
		t.Errorf("Tag is invalid, expected %s, got %s", version, ref.Tag)
	}
}
