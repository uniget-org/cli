package cache

import (
	"crypto/sha256"
	"encoding/hex"
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

func TestNewNoneCache(t *testing.T) {
	c := NewNoneCache()
	if c == nil {
		t.Errorf("Cache is invalid")
	}
}

func TestGet(t *testing.T) {
	ref := NewToolRef("127.0.0.1:5000", "uniget-org/tools", "test", "1.0.0")
	c := NewNoneCache()
	image, err := c.Get(ref)
	if err != nil {
		t.Errorf("Error getting image: %v", err)
	}
	if image == nil {
		t.Errorf("Image is invalid")
	}
	if len(image) == 0 {
		t.Errorf("Image is empty")
	}

	h := sha256.New()
	h.Write(image)
	bs := h.Sum(nil)
	if len(bs) == 0 {
		t.Errorf("Hash is empty")
	}
	sha256 := hex.EncodeToString(h.Sum(nil))
	if sha256 != "0f015b5bc195319dc5ae7eef10e4b7eb7903323793ed4b9461a38bf2948c64e6" {
		t.Errorf("Hash is invalid %s", sha256)
	}
}
