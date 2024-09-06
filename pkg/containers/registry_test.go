package containers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"testing"

	"github.com/regclient/regclient/types/blob"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
)

func TestGetRegclient(t *testing.T) {
	rc := GetRegclient()
	if rc == nil {
		t.Errorf("failed to get regclient")
	}
}

func TestGetImageTags(t *testing.T) {
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
	tags, err := GetImageTags(toolRef)
	if err != nil {
		t.Errorf("failed to get image tags: %s", err)
	}
	if len(tags) == 0 {
		t.Errorf("no tags found")
	}
}

func TestGetPlatformManifestOld(t *testing.T) {
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
	r, err := ref.New(toolRef.String())
	if err != nil {
		t.Errorf("failed to parse image name <%s>: %s", toolRef.String(), err)
	}
	rc := GetRegclient()
	defer rc.Close(context.Background(), r)

	m, err := GetPlatformManifestOld(context.Background(), rc, r)
	if err != nil {
		t.Errorf("failed to get platform manifest: %s", err)
	}
	if m == nil {
		t.Errorf("no platform manifest found")
	}
}

func TestGetManifestOld(t *testing.T) {
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)

	err := GetManifestOld(toolRef, func(blob blob.Reader) error {
		layer, err := io.ReadAll(blob)
		if err != nil {
			return fmt.Errorf("failed to read layer: %s", err)
		}
		if len(layer) == 0 {
			return fmt.Errorf("layer is empty")
		}
		return nil
	})
	if err != nil {
		t.Errorf("failed to get and process manifest: %s", err)
	}
}

func TestProcessLayersCallback(t *testing.T) {
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
	r, err := ref.New(toolRef.String())
	if err != nil {
		t.Errorf("failed to parse image name <%s>: %s", toolRef.String(), err)
	}
	rc := GetRegclient()
	defer rc.Close(context.Background(), r)

	m, err := GetPlatformManifestOld(context.Background(), rc, r)
	if err != nil {
		t.Errorf("failed to get platform manifest: %s", err)
	}

	ProcessLayersCallback(rc, m, r, func(blob blob.Reader) error {
		layer, err := io.ReadAll(blob)
		if err != nil {
			return fmt.Errorf("failed to read layer: %s", err)
		}
		if len(layer) == 0 {
			return fmt.Errorf("layer is empty")
		}
		return nil
	})
}

func TestGetPlatformManifestForLocalPlatform(t *testing.T) {
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
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
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
	expectedLayerSha := "0f015b5bc195319dc5ae7eef10e4b7eb7903323793ed4b9461a38bf2948c64e6"

	layerSha, err := GetFirstLayerShaFromRegistry(toolRef)
	if err != nil {
		t.Errorf("failed to get first layer sha: %s", err)
	}
	if layerSha != fmt.Sprintf("sha256:%s", expectedLayerSha) {
		t.Errorf("expected layer sha to be %s, got '%s'", expectedLayerSha, layerSha)
	}
}

func TestGetPlatformManifest(t *testing.T) {
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
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
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
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
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
	expectedLayerSha := "0f015b5bc195319dc5ae7eef10e4b7eb7903323793ed4b9461a38bf2948c64e6"

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
	if sha256 != expectedLayerSha {
		t.Errorf("Hash is invalid %s", sha256)
	}
}

func TestGetFirstLayerFromRegistry(t *testing.T) {
	toolRef := NewToolRef(
		"127.0.0.1:5000",
		"uniget-org/tools",
		"test",
		"1.0.0",
	)
	expectedLayerSha := "0f015b5bc195319dc5ae7eef10e4b7eb7903323793ed4b9461a38bf2948c64e6"

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
	if sha256 != expectedLayerSha {
		t.Errorf("Hash is invalid %s", sha256)
	}
}
