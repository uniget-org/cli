package containers

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"time"

	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
)

func GetRegclient() *regclient.RegClient {
	rcOpts := []regclient.Opt{}
	rcOpts = append(rcOpts, regclient.WithUserAgent("uniget"))
	rcOpts = append(rcOpts, regclient.WithDockerCreds())
	rcOpts = append(rcOpts, regclient.WithConfigHost(config.Host{
		Name: "127.0.0.1:5000",
		TLS:  config.TLSDisabled,
	}))

	return regclient.New(rcOpts...)
}

func GetImageTags(t *ToolRef) ([]string, error) {
	ctx := context.Background()

	r, err := ref.New(t.String())
	if err != nil {
		return []string{}, fmt.Errorf("failed to parse image name <%s>: %s", t.String(), err)
	}

	rc := GetRegclient()
	defer rc.Close(ctx, r)

	tags, err := rc.TagList(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %s", err)
	}

	var filteredTags []string
	for _, tag := range tags.Tags {
		if tag == "latest" || tag == "main" || tag == "test" {
			continue
		}

		filteredTags = append(filteredTags, tag)
	}

	return filteredTags, nil
}

func GetFirstLayerShaFromRegistry(image *ToolRef) (string, error) {
	ctx := context.Background()

	r, err := ref.New(image.String())
	if err != nil {
		return "", fmt.Errorf("failed to parse image name <%s>: %s", image, err)
	}

	rc := GetRegclient()
	defer rc.Close(ctx, r)

	manifestCtx, manifestCancel := context.WithTimeout(ctx, 60*time.Second)
	defer manifestCancel()
	m, err := GetPlatformManifestForLocalPlatform(manifestCtx, rc, r)
	if err != nil {
		return "", fmt.Errorf("failed to get manifest: %s", err)
	}

	if m.IsList() {
		return "", fmt.Errorf("manifest is a list")
	}

	mi, ok := m.(manifest.Imager)
	if !ok {
		return "", fmt.Errorf("failed to get imager")
	}

	layers, err := mi.GetLayers()
	if err != nil {
		return "", fmt.Errorf("failed to get layers: %s", err)
	}

	if len(layers) > 1 {
		return "", fmt.Errorf("image must have exactly one layer but got %d", len(layers))
	}

	layer := layers[0]
	if layer.MediaType == mediatype.OCI1Layer || layer.MediaType == mediatype.OCI1LayerZstd {
		return "", fmt.Errorf("only layers with gzip compression are supported (not %s)", layer.MediaType)
	}
	if layer.MediaType == mediatype.OCI1LayerGzip || layer.MediaType == mediatype.Docker2LayerGzip {

		return string(layer.Digest), nil
	}

	return "", fmt.Errorf("unknown media type encountered: %s", layer.MediaType)
}

func HeadPlatformManifestForLocalPlatform(ctx context.Context, rc *regclient.RegClient, r ref.Ref) (bool, error) {
	return HeadPlatformManifest(ctx, rc, r, platform.Local())
}

func GetPlatformManifestForLocalPlatform(ctx context.Context, rc *regclient.RegClient, r ref.Ref) (manifest.Manifest, error) {
	return GetPlatformManifest(ctx, rc, r, platform.Local())
}

func HeadPlatformManifest(ctx context.Context, rc *regclient.RegClient, r ref.Ref, p platform.Platform) (bool, error) {
	m, err := rc.ManifestHead(ctx, r)
	if err != nil {
		return false, fmt.Errorf("failed to get manifest: %s", err)
	}

	if m.IsList() {
		desc, err := manifest.GetPlatformDesc(m, &p)
		if err != nil {
			return false, fmt.Errorf("error getting platform descriptor")
		}

		_, err = rc.ManifestHead(ctx, r, regclient.WithManifestDesc(*desc))
		if err != nil {
			return false, fmt.Errorf("failed to get manifest: %s", err)
		}
	}

	return true, nil
}

func GetPlatformManifest(ctx context.Context, rc *regclient.RegClient, r ref.Ref, p platform.Platform) (manifest.Manifest, error) {
	m, err := rc.ManifestGet(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %s", err)
	}

	if m.IsList() {
		desc, err := manifest.GetPlatformDesc(m, &p)
		if err != nil {
			return nil, fmt.Errorf("error getting platform descriptor")
		}

		_, err = rc.ManifestGet(ctx, r, regclient.WithManifestDesc(*desc))
		if err != nil {
			return nil, fmt.Errorf("failed to get manifest: %s", err)
		}
	}

	return m, nil
}

func HeadManifest(ctx context.Context, rc *regclient.RegClient, r ref.Ref) (bool, error) {
	manifestCtx, manifestCancel := context.WithTimeout(ctx, 60*time.Second)
	defer manifestCancel()

	_, err := HeadPlatformManifestForLocalPlatform(manifestCtx, rc, r)
	if err != nil {
		return false, fmt.Errorf("failed to get manifest: %s", err)
	}

	return true, nil
}

func GetManifest(ctx context.Context, rc *regclient.RegClient, r ref.Ref) (manifest.Manifest, error) {
	manifestCtx, manifestCancel := context.WithTimeout(ctx, 60*time.Second)
	defer manifestCancel()

	m, err := GetPlatformManifestForLocalPlatform(manifestCtx, rc, r)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %s", err)
	}

	return m, nil
}

func GetFirstLayerFromManifest(ctx context.Context, rc *regclient.RegClient, m manifest.Manifest) ([]byte, error) {
	return GetLayerFromManifestByIndex(ctx, rc, m, 0)
}

func GetLayerFromManifestByIndex(ctx context.Context, rc *regclient.RegClient, m manifest.Manifest, index int) ([]byte, error) {
	if m.IsList() {
		return nil, fmt.Errorf("manifest is a list")
	}

	mi, ok := m.(manifest.Imager)
	if !ok {
		return nil, fmt.Errorf("failed to get imager")
	}

	layers, err := mi.GetLayers()
	if err != nil {
		return nil, fmt.Errorf("failed to get layers: %s", err)
	}

	if len(layers) < index {
		return nil, fmt.Errorf("image only has %d layers", len(layers))
	}

	layer := layers[index]
	if layer.MediaType == mediatype.OCI1Layer || layer.MediaType == mediatype.OCI1LayerZstd {
		return nil, fmt.Errorf("only layers with gzip compression are supported (not %s)", layer.MediaType)
	}
	if layer.MediaType == mediatype.OCI1LayerGzip || layer.MediaType == mediatype.Docker2LayerGzip {

		d, err := digest.Parse(string(layer.Digest))
		if err != nil {
			return nil, fmt.Errorf("failed to parse digest %s: %s", layer.Digest, err)
		}

		blob, err := rc.BlobGet(context.Background(), m.GetRef(), descriptor.Descriptor{Digest: d})
		if err != nil {
			return nil, fmt.Errorf("failed to get blob for digest %s: %s", layer.Digest, err)
		}
		defer blob.Close()

		layerData, err := blob.RawBody()
		if err != nil {
			return nil, fmt.Errorf("failed to read layer: %s", err)
		}

		return layerData, nil
	}

	return nil, fmt.Errorf("unsupported layer media type %s", layer.MediaType)
}

func gunzip(layer []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(layer))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %s", err)
	}
	defer reader.Close()

	buffer, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzip: %s", err)
	}

	return buffer, nil
}

func GetFirstLayerFromRegistry(ctx context.Context, rc *regclient.RegClient, r ref.Ref) ([]byte, error) {
	m, err := GetManifest(ctx, rc, r)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %s", err)
	}

	imageGz, err := GetFirstLayerFromManifest(ctx, rc, m)
	if err != nil {
		return nil, fmt.Errorf("failed to get first layer: %s", err)
	}

	image, err := gunzip(imageGz)
	if err != nil {
		return nil, fmt.Errorf("failed to gunzip layer: %s", err)
	}

	return image, nil
}
