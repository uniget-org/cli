package containers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
)

func TestGetRegclient(t *testing.T) {
	rc := GetRegclient()
	if rc == nil {
		t.Errorf("failed to get regclient")
	}
}

func TestGetImageTagsInvalidImage(t *testing.T) {
	registryAddress := "127.0.0.1:5000:5001"
	toolRef := NewToolRef(registryAddress, "a/b", "c", "d+e")
	_, err := GetImageTags(toolRef)
	if err == nil {
		t.Errorf("expected error due to invalid registry host %s: %s", registryAddress, err)
	}
}

func TestGetImageTagsUnreachableRegistry(t *testing.T) {
	registryAddress := "127.0.0.1:5001"
	toolRef := NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	_, err := GetImageTags(toolRef)
	if err == nil {
		t.Errorf("expected error due to unreachable registry %s: %s", registryAddress, err)
	}
}

func TestGetImageTags(t *testing.T) {
	tags, err := GetImageTags(toolRef)
	if err != nil {
		t.Errorf("failed to get image tags: %s", err)
	}
	if len(tags) == 0 {
		t.Errorf("no tags found")
	}
}

func TestGetPlatformManifestForLocalPlatform(t *testing.T) {
	toolRef = NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	r, err := ref.New(toolRef.String())
	if err != nil {
		t.Errorf("failed to parse image name <%s>: %s", toolRef.String(), err)
	}
	rc := GetRegclient()
	defer rc.Close(context.Background(), r)

	m, err := GetPlatformManifestForLocalPlatform(context.Background(), rc, r)
	if err != nil {
		t.Errorf("failed to get platform manifest: %s", err)
	}
	if m == nil {
		t.Errorf("no platform manifest found")
	}
}

func TestGetFirstLayerShaFromRegistry(t *testing.T) {
	toolRef = NewToolRef(registryAddress, registryRepository, registryImage, registryTag)

	layerSha, err := GetFirstLayerShaFromRegistry(toolRef)
	if err != nil {
		t.Errorf("failed to get first layer sha: %s", err)
	}
	if layerSha != fmt.Sprintf("sha256:%s", expectedLayerGzSha256) {
		t.Errorf("expected layer sha to be %s, got '%s'", expectedLayerGzSha256, layerSha)
	}
}

func TestGetPlatformManifest(t *testing.T) {
	toolRef = NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	r, err := ref.New(toolRef.String())
	if err != nil {
		t.Errorf("failed to parse image name <%s>: %s", toolRef.String(), err)
	}
	rc := GetRegclient()
	defer rc.Close(context.Background(), r)

	m, err := GetPlatformManifest(context.Background(), rc, r, platform.Local())
	if err != nil {
		t.Errorf("failed to get platform manifest: %s", err)
	}
	if m == nil {
		t.Errorf("no platform manifest found")
	}
}

func TestGetManifest(t *testing.T) {
	toolRef = NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	r, err := ref.New(toolRef.String())
	if err != nil {
		t.Errorf("failed to parse image name <%s>: %s", toolRef.String(), err)
	}
	rc := GetRegclient()
	defer rc.Close(context.Background(), r)

	m, err := GetManifest(context.Background(), rc, r)
	if err != nil {
		t.Errorf("failed to get platform manifest: %s", err)
	}
	if m == nil {
		t.Errorf("no platform manifest found")
	}
}

func TestGetFirstLayerFromManifest(t *testing.T) {
	toolRef = NewToolRef(registryAddress, registryRepository, registryImage, registryTag)

	r, err := ref.New(toolRef.String())
	if err != nil {
		t.Errorf("failed to parse image name <%s>: %s", toolRef.String(), err)
	}
	rc := GetRegclient()
	defer rc.Close(context.Background(), r)

	m, err := GetPlatformManifest(context.Background(), rc, r, platform.Local())
	if err != nil {
		t.Errorf("failed to get platform manifest: %s", err)
	}

	layer, err := GetFirstLayerFromManifest(context.Background(), rc, m)
	if err != nil {
		t.Errorf("failed to get platform manifest: %s", err)
	}
	if len(layer) == 0 {
		t.Errorf("layer is empty")
	}

	h := sha256.New()
	h.Write(layer)
	bs := h.Sum(nil)
	if len(bs) == 0 {
		t.Errorf("Hash is empty")
	}
	sha256 := hex.EncodeToString(h.Sum(nil))
	if sha256 != expectedLayerGzSha256 {
		t.Errorf("Hash is invalid %s", sha256)
	}
}

func TestGetFirstLayerFromRegistry(t *testing.T) {
	toolRef = NewToolRef(registryAddress, registryRepository, registryImage, registryTag)

	r, err := ref.New(toolRef.String())
	if err != nil {
		t.Errorf("failed to parse image name <%s>: %s", toolRef.String(), err)
	}

	rc := GetRegclient()
	defer rc.Close(context.Background(), r)

	layer, err := GetFirstLayerFromRegistry(context.Background(), rc, r)
	if err != nil {
		t.Errorf("failed to get first layer: %s", err)
	}
	if len(layer) == 0 {
		t.Errorf("layer is empty")
	}

	h := sha256.New()
	h.Write(layer)
	bs := h.Sum(nil)
	if len(bs) == 0 {
		t.Errorf("Hash is empty")
	}
	sha256 := hex.EncodeToString(h.Sum(nil))
	if sha256 != expectedLayerGzSha256 {
		t.Errorf("expected layer Sha256 %s but got %s", expectedLayerGzSha256, sha256)
	}
}
