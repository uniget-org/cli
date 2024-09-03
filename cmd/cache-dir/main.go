package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
)

func GetPlatformManifestForLocalPlatform(ctx context.Context, rc *regclient.RegClient, r ref.Ref) (manifest.Manifest, error) {
	return GetPlatformManifest(ctx, rc, r, platform.Local())
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

		m, err = rc.ManifestGet(ctx, r, regclient.WithManifestDesc(*desc))
		if err != nil {
			return nil, fmt.Errorf("failed to get manifest: %s", err)
		}
	}

	return m, nil
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

func GetFirstLayerFromManifest(ctx context.Context, rc *regclient.RegClient, m manifest.Manifest) ([] byte, error) {
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

func GetFirstLayerFromRegistry(ctx context.Context, rc *regclient.RegClient, r ref.Ref) ([]byte, error) {
	m, err := GetManifest(ctx, rc, r)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %s", err)
	}

	return GetFirstLayerFromManifest(ctx, rc, m)
}

func WriteDataToCache(data []byte, key string) error {
	err := os.WriteFile(fmt.Sprintf("%s/%s", cacheDirectory, key), data, 0644) // #nosec G306 -- just for testing
	if err != nil {
		return fmt.Errorf("failed to write data for key %s to cache: %s", key, err)
	}
	return nil
}

func CheckDataInCache(key string) bool {
	_, err := os.Stat(fmt.Sprintf("%s/%s", cacheDirectory, key))
	return !os.IsNotExist(err)
}

func ReadDataFromCache(key string) ([]byte, error) {
	data, err := os.ReadFile(fmt.Sprintf("%s/%s", cacheDirectory, key))
	if err != nil {
		return nil, fmt.Errorf("failed to read data for key %s from cache: %s", key, err)
	}
	return data, nil
}

var (
	projectName = "uniget"

	registry = "ghcr.io"
	repository = "uniget-org/tools"
	tool = "jq"
	version = "1.7.1"

	cacheRoot string
	cacheDirectory string
)

func main() {
	toolRef := fmt.Sprintf("%s/%s/%s:%s", registry, repository, tool, version)

	user := true
	if user {
		cacheRoot = os.Getenv("HOME") + "/.cache"
		if os.Getenv("XDG_CACHE_HOME") != "" {
			if strings.HasPrefix(os.Getenv("XDG_CACHE_HOME"), os.Getenv("HOME")) {
				cacheRoot = os.Getenv("XDG_CACHE_HOME")
			}
		}

	} else {
		cacheRoot = "/var/cache"
	}
	cacheDirectory = cacheRoot + "/" + projectName + "/download"
	err := os.MkdirAll(cacheDirectory, 0755) // #nosec G301 -- cache directory
	if err != nil {
		panic(err)
	}

	cacheKey := fmt.Sprintf("%s-%s", tool, version)
	if CheckDataInCache(cacheKey) {
		fmt.Printf("Cache hit for %s\n", cacheKey)

	} else {
		fmt.Printf("Cache miss for %s\n", cacheKey)

		ctx := context.Background()

		r, err := ref.New(toolRef)
		if err != nil {
			panic(err)
		}

		rcOpts := []regclient.Opt{}
		rcOpts = append(rcOpts, regclient.WithUserAgent("uniget"))
		rcOpts = append(rcOpts, regclient.WithDockerCreds())
		rc := regclient.New(rcOpts...)
		defer rc.Close(ctx, r)

		layer, err := GetFirstLayerFromRegistry(ctx, rc, r)
		if err != nil {
			panic(err)
		}

		err = WriteDataToCache(layer, cacheKey)
		if err != nil {
			panic(err)
		}
	}

	layer, err := ReadDataFromCache(cacheKey)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(fmt.Sprintf("%s-%s.tar", tool, version), layer, 0644) // #nosec G306 -- just for testing
	if err != nil {
		panic(err)
	}
}