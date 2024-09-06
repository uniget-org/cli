package cache

import (
	"testing"

	"github.com/uniget-org/cli/pkg/containers"
)

func TestNewToolRef(t *testing.T) {
	ref := containers.NewToolRef("a", "b", "c", "d")

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
	ref := containers.NewToolRef("a", "b", "c", "d")

	if ref.String() != "a/b/c:d" {
		t.Errorf("String is invalid")
	}
}
