package containers

import (
	"context"
	"fmt"

	"github.com/regclient/regclient"
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