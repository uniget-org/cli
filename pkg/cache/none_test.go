package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"
)

func TestNewNoneCache(t *testing.T) {
	c := NewNoneCache()
	if c == nil {
		t.Errorf("Cache is invalid")
	}
}

func TestNoneCacheGet(t *testing.T) {
	c := NewNoneCache()
	imageReader, err := c.Get(toolRef)
	if err != nil {
		t.Errorf("Error getting image: %v", err)
	}
	image, err := io.ReadAll(imageReader)
	if err != nil {
		t.Errorf("Error reading image: %v", err)
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
	if sha256 != expectedLayerSha256 {
		t.Errorf("expected sha256 %s but got %s", expectedLayerSha256, sha256)
	}
}
