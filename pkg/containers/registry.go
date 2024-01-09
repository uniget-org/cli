package containers

import (
	"context"
	"fmt"
	"time"

	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/blob"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/ref"
)

func GetPlatformManifest(ctx context.Context, rc *regclient.RegClient, r ref.Ref, altArch string) (manifest.Manifest, error) {
	m, err := rc.ManifestGet(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %s", err)
	}

	if m.IsList() {

		mi, ok := m.(manifest.Indexer)
		if !ok {
			return nil, fmt.Errorf("failed to get imager")
		}
		manifests, err := mi.GetManifestList()
		if err != nil {
			return nil, fmt.Errorf("error getting manifests")
		}

		for _, manifest := range manifests {

			if manifest.Platform.Architecture == altArch {
				platformImage := fmt.Sprintf("%s@%s", r.Reference, manifest.Digest)
				r2, err := ref.New(platformImage)
				if err != nil {
					return nil, fmt.Errorf("failed to parse image name <%s>: %s", platformImage, err)
				}

				m, err := rc.ManifestGet(ctx, r2)
				if err != nil {
					return nil, fmt.Errorf("failed to get manifest: %s", err)
				}

				if m.IsList() {
					return nil, fmt.Errorf("manifest cannot be list again")
				}

				return m, nil
			}
		}
	}

	return m, nil
}

func GetManifest(image string, altArch string, callback func(blob blob.Reader) error) error {
	ctx := context.Background()

	r, err := ref.New(image)
	if err != nil {
		return fmt.Errorf("failed to parse image name <%s>: %s", image, err)
	}

	rcOpts := []regclient.Opt{}
	rcOpts = append(rcOpts, regclient.WithUserAgent("uniget"))
	rcOpts = append(rcOpts, regclient.WithDockerCreds())
	rc := regclient.New(rcOpts...)
	defer rc.Close(ctx, r)

	manifestCtx, manifestCancel := context.WithTimeout(ctx, 60*time.Second)
	defer manifestCancel()
	m, err := GetPlatformManifest(manifestCtx, rc, r, altArch)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %s", err)
	}

	err = ProcessLayersCallback(ctx, rc, m, r, callback)
	if err != nil {
		return fmt.Errorf("failed to process layers with callback: %s", err)
	}

	return nil
}

func ProcessLayersCallback(ctx context.Context, rc *regclient.RegClient, m manifest.Manifest, r ref.Ref, callback func(blob blob.Reader) error) error {
	if m.IsList() {
		return fmt.Errorf("manifest is a list")
	}

	mi, ok := m.(manifest.Imager)
	if !ok {
		return fmt.Errorf("failed to get imager")
	}

	layers, err := mi.GetLayers()
	if err != nil {
		return fmt.Errorf("failed to get layers: %s", err)
	}

	if len(layers) > 1 {
		return fmt.Errorf("image must have exactly one layer but got %d", len(layers))
	}

	layer := layers[0]
	if layer.MediaType == types.MediaTypeOCI1Layer || layer.MediaType == types.MediaTypeOCI1LayerZstd {
		return fmt.Errorf("only layers with gzip compression are supported (not %s)", layer.MediaType)
	}
	if layer.MediaType == types.MediaTypeOCI1LayerGzip || layer.MediaType == types.MediaTypeDocker2LayerGzip {

		d, err := digest.Parse(string(layer.Digest))
		if err != nil {
			return fmt.Errorf("failed to parse digest %s: %s", layer.Digest, err)
		}

		layerCtx, layerCancel := context.WithTimeout(ctx, 180*time.Second)
		defer layerCancel()
		blob, err := rc.BlobGet(layerCtx, r, types.Descriptor{Digest: d})
		if err != nil {
			return fmt.Errorf("failed to get blob for digest %s: %s", layer.Digest, err)
		}
		defer blob.Close()

		err = callback(blob)
		if err != nil {
			return fmt.Errorf("failed callback: %s", err)
		}

		return nil
	}

	return fmt.Errorf("unknown media type encountered: %s", layer.MediaType)
}
