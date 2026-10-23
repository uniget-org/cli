package containers

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/platform"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

// cloneDecoded returns a fresh copy of a manifest decoded into a map.
func cloneDecoded(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("failed to decode manifest: %s", err)
	}
	return decoded
}

func layersOf(t *testing.T, decoded map[string]any) []any {
	t.Helper()
	layers, ok := decoded["layers"].([]any)
	if !ok || len(layers) == 0 {
		t.Fatalf("expected manifest layers, got %v", decoded["layers"])
	}
	return layers
}

// putManifest pushes raw JSON as a new tag and returns the reference.
func putManifest(t *testing.T, mediaType string, raw []byte, suffix string) *ToolRef {
	t.Helper()
	desc := descriptor.Descriptor{
		MediaType: mediaType,
		Digest:    digest.FromBytes(raw),
		Size:      int64(len(raw)),
	}
	m, err := manifest.New(manifest.WithDesc(desc), manifest.WithRaw(raw))
	if err != nil {
		t.Fatalf("failed to build synthetic manifest: %s", err)
	}

	toolRef := NewToolRef(registryAddress, registryRepository, registryImage, "synthetic-"+suffix)
	rc := GetRegclient()
	defer func() {
		_ = rc.Close(context.Background(), toolRef.GetRef())
	}()

	err = rc.ManifestPut(context.Background(), toolRef.GetRef(), m)
	if err != nil {
		t.Fatalf("failed to push synthetic manifest %s: %s", toolRef.String(), err)
	}
	return toolRef
}

// buildPlainManifest pushes a manifest derived from the jq image where the
// layer descriptors may be duplicated or have modified media types.
func buildPlainManifest(t *testing.T, mutate func(t *testing.T, layers []any) []any, suffix string) *ToolRef {
	t.Helper()

	base := NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	rc := GetRegclient()
	defer func() {
		_ = rc.Close(context.Background(), base.GetRef())
	}()

	m, err := GetPlatformManifestForLocalPlatform(context.Background(), rc, base.GetRef())
	if err != nil {
		t.Fatalf("failed to get base manifest: %s", err)
	}
	baseRaw, err := m.RawBody()
	if err != nil {
		t.Fatalf("failed to get raw manifest: %s", err)
	}
	decoded := cloneDecoded(t, baseRaw)
	layers := layersOf(t, decoded)
	decoded["layers"] = mutate(t, layers)

	raw, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("failed to encode synthetic manifest: %s", err)
	}
	return putManifest(t, m.GetDescriptor().MediaType, raw, suffix)
}

func TestUncoveredRegistryFunctions(t *testing.T) {
	unreachable := NewToolRef("127.0.0.1:5001", registryRepository, registryImage, registryTag).GetRef()

	digest, err := GetImageDigest(toolRef)
	if err != nil {
		t.Errorf("failed to get image digest: %s", err)
	}
	if digest == "" {
		t.Errorf("expected digest, got empty")
	}

	labels, err := GetImageLabels(toolRef)
	if err != nil {
		t.Errorf("failed to get image labels: %s", err)
	}
	if labels == nil {
		t.Errorf("expected labels map, got nil")
	}

	_, err = GetImageDigest(NewToolRef("127.0.0.1:5000:5000", "a", "b", "c+d"))
	if err == nil {
		t.Errorf("expected error for invalid reference, got nil")
	}

	_, err = GetImageLabels(unreachableToToolRef("127.0.0.1:5001"))
	if err == nil {
		t.Errorf("expected error for unreachable registry, got nil")
	}

	rc := GetRegclient()
	defer func() {
		_ = rc.Close(context.Background(), toolRef.GetRef())
	}()

	exists, err := HeadPlatformManifestForLocalPlatform(context.Background(), rc, toolRef.GetRef())
	if err != nil || !exists {
		t.Errorf("expected manifest to exist, got %v err %v", exists, err)
	}

	_, err = HeadPlatformManifestForLocalPlatform(context.Background(), rc, unreachable)
	if err == nil {
		t.Errorf("expected error for unreachable registry, got nil")
	}

	exists, err = HeadPlatformManifest(context.Background(), rc, toolRef.GetRef(), platform.Local())
	if err != nil || !exists {
		t.Errorf("expected manifest to exist, got %v err %v", exists, err)
	}

	exists, err = HeadManifest(context.Background(), rc, toolRef.GetRef())
	if err != nil || !exists {
		t.Errorf("expected manifest to exist, got %v err %v", exists, err)
	}

	_, err = HeadManifest(context.Background(), rc, unreachable)
	if err == nil {
		t.Errorf("expected error for unreachable registry, got nil")
	}

	_, err = GetManifest(context.Background(), rc, unreachable)
	if err == nil {
		t.Errorf("expected error for unreachable registry, got nil")
	}
}

func unreachableToToolRef(host string) *ToolRef {
	return NewToolRef(host, registryRepository, registryImage, registryTag)
}

func TestManifestExistsWithLocalRegistry(t *testing.T) {
	existing := NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	if !existing.ManifestExists() {
		t.Errorf("expected manifest to exist for %s", existing.String())
	}

	missing := NewToolRef(registryAddress, registryRepository, "nonexistent-tool", "9.9.9")
	if missing.ManifestExists() {
		t.Errorf("expected manifest to not exist for %s", missing.String())
	}
}

func TestFindToolRefWithLocalRegistry(t *testing.T) {
	registries := []string{registryAddress, "127.0.0.1:5001"}
	repositories := []string{registryRepository, registryRepository}

	found, err := FindToolRef(registries, repositories, registryImage, registryTag)
	if err != nil {
		t.Fatalf("expected tool reference to be found, got %s", err)
	}
	if found.Registry != registryAddress {
		t.Errorf("expected registry %s, got %s", registryAddress, found.Registry)
	}

	_, err = FindToolRef(registries, repositories, "nonexistent-tool", "9.9.9")
	if err == nil {
		t.Errorf("expected error for missing tool, got nil")
	}
}

func TestFindToolRefErrors(t *testing.T) {
	tests := []struct {
		name         string
		registries   []string
		repositories []string
	}{
		{
			name:         "no registries",
			registries:   []string{},
			repositories: []string{"repository"},
		},
		{
			name:         "no repositories",
			registries:   []string{"registry"},
			repositories: []string{},
		},
		{
			name:         "count mismatch",
			registries:   []string{"registry"},
			repositories: []string{"repository", "repository"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FindToolRef(tt.registries, tt.repositories, "tool", "1.0.0")
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestRepositoryString(t *testing.T) {
	ref := NewToolRef("ghcr.io", "uniget-org/tools", "jq", "1.7.1")
	if got := ref.RepositoryString(); got != "ghcr.io/uniget-org/tools/jq" {
		t.Errorf("expected ghcr.io/uniget-org/tools/jq, got %s", got)
	}
}

func TestGetRefWithInvalidName(t *testing.T) {
	ref := NewToolRef("this is not valid", "b", "c", "d")
	r := ref.GetRef()
	if r.Repository != "" {
		t.Errorf("expected empty ref for invalid image name, got %v", r.Repository)
	}
}

func TestGetFirstLayerShaFromRegistryRejectsBadShapes(t *testing.T) {
	unreachable := NewToolRef("127.0.0.1:5001", registryRepository, registryImage, registryTag)
	if _, err := GetFirstLayerShaFromRegistry(unreachable); err == nil {
		t.Errorf("expected error for unreachable registry, got nil")
	}
	if _, err := GetFirstLayerShaFromRegistry(NewToolRef("127.0.0.1:5000:5000", "a", "b", "c+d")); err == nil {
		t.Errorf("expected error for invalid reference, got nil")
	}

	multiLayer := buildPlainManifest(t, func(_ *testing.T, layers []any) []any {
		return []any{layers[0], layers[0]}
	}, "twolayers")
	if _, err := GetFirstLayerShaFromRegistry(multiLayer); err == nil || !strings.Contains(err.Error(), "exactly one layer") {
		t.Errorf("expected multi layer error, got %v", err)
	}

	zstd := buildPlainManifest(t, func(_ *testing.T, layers []any) []any {
		entry := layers[0].(map[string]any)
		entry["mediaType"] = "application/vnd.oci.image.layer.v1.tar+zstd"
		return []any{entry}
	}, "zstd")
	if _, err := GetFirstLayerShaFromRegistry(zstd); err == nil || !strings.Contains(err.Error(), "gzip compression") {
		t.Errorf("expected zstd error, got %v", err)
	}

	unknown := buildPlainManifest(t, func(_ *testing.T, layers []any) []any {
		entry := layers[0].(map[string]any)
		entry["mediaType"] = "application/vnd.test.unknown"
		return []any{entry}
	}, "unknown")
	if _, err := GetFirstLayerShaFromRegistry(unknown); err == nil || !strings.Contains(err.Error(), "unknown media type") {
		t.Errorf("expected unknown media type error, got %v", err)
	}
}

func TestGetLayerFromManifestByIndexRejectsBadShapes(t *testing.T) {
	rc := GetRegclient()
	defer func() {
		_ = rc.Close(context.Background(), toolRef.GetRef())
	}()
	progress := tui.NewProgressReader(nil, nil)
	noop := func(_ io.ReadCloser) error { return nil }

	platformResolved, err := GetManifest(context.Background(), rc, toolRef.GetRef())
	if err != nil {
		t.Fatalf("failed to get manifest: %s", err)
	}
	err = GetLayerFromManifestByIndex(context.Background(), rc, platformResolved, 10, progress, noop)
	if err == nil || !strings.Contains(err.Error(), "image only has") {
		t.Errorf("expected out of range error, got %v", err)
	}

	zstd := buildPlainManifest(t, func(_ *testing.T, layers []any) []any {
		entry := layers[0].(map[string]any)
		entry["mediaType"] = "application/vnd.oci.image.layer.v1.tar+zstd"
		return []any{entry}
	}, "zstd2")
	zstdResolved, err := GetManifest(context.Background(), rc, zstd.GetRef())
	if err != nil {
		t.Fatalf("failed to get manifest: %s", err)
	}
	err = GetLayerFromManifestByIndex(context.Background(), rc, zstdResolved, 0, progress, noop)
	if err == nil || !strings.Contains(err.Error(), "gzip compression") {
		t.Errorf("expected zstd error, got %v", err)
	}

	unknown := buildPlainManifest(t, func(_ *testing.T, layers []any) []any {
		entry := layers[0].(map[string]any)
		entry["mediaType"] = "application/vnd.test.unknown.2"
		return []any{entry}
	}, "unknown2")
	unknownResolved, err := GetManifest(context.Background(), rc, unknown.GetRef())
	if err != nil {
		t.Fatalf("failed to get manifest: %s", err)
	}
	err = GetLayerFromManifestByIndex(context.Background(), rc, unknownResolved, 0, progress, noop)
	if err == nil || !strings.Contains(err.Error(), "unsupported layer") {
		t.Errorf("expected unsupported media type error, got %v", err)
	}
}

func TestVerifyContainerImageSignature(t *testing.T) {
	bad := NewToolRef("127.0.0.1:5000:5000", "a", "b", "c+d")
	if err := VerifyContainerImageSignature(bad, "", "", "", ""); err == nil {
		t.Errorf("expected error for invalid reference, got nil")
	}

	err := VerifyContainerImageSignature(toolRef, "https://invalid-issuer.test", "", "invalid@test.example", "")
	if err == nil {
		t.Errorf("expected signature verification error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "vfor") {
		t.Errorf("expected signature verification error, got %v", err)
	}
}
