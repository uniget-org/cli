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
	"github.com/regclient/regclient/types/ref"
	"github.com/regclient/regclient/types/manifest"
)

func GetPlatformManifest(ctx context.Context, rc *regclient.RegClient, r ref.Ref, alt_arch string) (manifest.Manifest, error) {
	m, err := rc.ManifestGet(ctx, r)
	if err != nil {
        return nil, fmt.Errorf("Failed to get manifest: %s\n", err)
	}

	// TODO: Test manifest list with Docker media types
	// TODO: Test manifest list with OCI media types
	// TODO: Test image with Docker media types
	// TODO: Test image with OCI media types
	if m.IsList() {

		mi, ok := m.(manifest.Indexer)
		if !ok {
			return nil, fmt.Errorf("ERROR")
		}
		manifests, err := mi.GetManifestList()
		if err != nil {
			return nil, fmt.Errorf("Error getting manifests")
		}

		for _, manifest := range manifests {

			if manifest.Platform.Architecture == alt_arch {
				platformImage := fmt.Sprintf("%s@%s", r.Reference, manifest.Digest)
				r2, err := ref.New(platformImage)
				if err != nil {
					return nil, fmt.Errorf("Failed to parse image name <%s>: %s\n", platformImage, err)
				}

				m, err := rc.ManifestGet(ctx, r2)
				if err != nil {
					return nil, fmt.Errorf("Failed to get manifest: %s\n", err)
				}

				if m.IsList() {
					return nil, fmt.Errorf("Manifest cannot be list again")
				}

				return m, nil
			}
		}
	}

	return m, nil
}

func GetManifest(image string, alt_arch string, callback func(blob blob.Reader) error) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30 * time.Second))
	defer cancel()

	r, err := ref.New(image)
	if err != nil {
		return fmt.Errorf("Failed to parse image name <%s>: %s\n", image, err)
	}

	rcOpts := []regclient.Opt{}
	rcOpts = append(rcOpts, regclient.WithUserAgent("docker-setup"))
	rcOpts = append(rcOpts, regclient.WithDockerCreds())
	rc := regclient.New(rcOpts...)
	defer rc.Close(ctx, r)

	m, err := GetPlatformManifest(ctx, rc, r, alt_arch)
	if err != nil {
		return fmt.Errorf("Failed to get manifest: %s\n", err)
	}
	err = ProcessLayersCallback(ctx, rc, m, r, callback)
	if err != nil {
		return fmt.Errorf("Failed to process layers with callback: %s\n", err)
	}

	return nil
}

func ProcessLayersCallback(ctx context.Context, rc *regclient.RegClient, m manifest.Manifest, r ref.Ref, callback func(blob blob.Reader) error) error {
	if m.IsList() {
		return fmt.Errorf("Manifest is a list")
	}

	mi, ok := m.(manifest.Imager)
	if !ok {
		return fmt.Errorf("ERROR")
	}

	layers, err := mi.GetLayers()
	if err != nil {
		return fmt.Errorf("Failed to get layers: %s", err)
	}
	
	if len(layers) > 1 {
		return fmt.Errorf("Image must have exactly one layer but got %d", len(layers))
	}

	layer := layers[0]
	// TODO: Test known but unsupported media types
	if layer.MediaType == types.MediaTypeOCI1Layer || layer.MediaType == types.MediaTypeOCI1LayerZstd {
		return fmt.Errorf("Only layers with gzip compression are supported (not %s)", layer.MediaType)
	}
	if layer.MediaType == types.MediaTypeOCI1LayerGzip || layer.MediaType == types.MediaTypeDocker2LayerGzip  {
		
		d, err := digest.Parse(string(layer.Digest))
		if err != nil {
			return fmt.Errorf("Failed to parse digest %s: %s", layer.Digest, err)
		}

		blob, err := rc.BlobGet(ctx, r, types.Descriptor{Digest: d})
		if err != nil {
			return fmt.Errorf("Failed to get blob for digest %s: %s", layer.Digest, err)
		}
		defer blob.Close()

		//fmt.Printf("len of blob: %d\n", len(blob))
		//fmt.Printf("type of blob: %T\n", blob)

		err = callback(blob)
		if err != nil {
			return fmt.Errorf("Failed callback: %s", err)
		}

		return nil
	}
	
	// TODO: Test unknown media types
	return fmt.Errorf("Unknown media type encountered: %s", layer.MediaType)
}