package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/uniget-org/cli/pkg/containers"
)

func TestNewNoneCache(t *testing.T) {
	c := NewNoneCache()
	if c == nil {
		t.Errorf("Cache is invalid")
	}
}

func TestNoneCacheGet(t *testing.T) {
	ref := containers.NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
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
