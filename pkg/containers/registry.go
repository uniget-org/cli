package containers

import (
	"context"
	"fmt"
	"time"

	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/opencontainers/go-digest"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/blob"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/platform"
	"github.com/regclient/regclient/types/ref"
)

func GetImageTags(repository string) ([]string, error) {
	ctx := context.Background()

	r, err := ref.New(repository)
	if err != nil {
		return []string{}, fmt.Errorf("failed to parse image name <%s>: %s", repository, err)
	}

	rcOpts := []regclient.Opt{}
	rcOpts = append(rcOpts, regclient.WithUserAgent("uniget"))
	rcOpts = append(rcOpts, regclient.WithDockerCreds())
	rc := regclient.New(rcOpts...)
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

func GetPlatformManifest(ctx context.Context, rc *regclient.RegClient, r ref.Ref) (manifest.Manifest, error) {
	m, err := rc.ManifestGet(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %s", err)
	}

	if m.IsList() {
		plat := platform.Local()
		desc, err := manifest.GetPlatformDesc(m, &plat)
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

func GetManifest(image string, callback func(blob blob.Reader) error) error {
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
	m, err := GetPlatformManifest(manifestCtx, rc, r)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %s", err)
	}

	err = ProcessLayersCallback(rc, m, r, callback)
	if err != nil {
		return fmt.Errorf("failed to process layers with callback: %s", err)
	}

	return nil
}

func ProcessLayersCallback(rc *regclient.RegClient, m manifest.Manifest, r ref.Ref, callback func(blob blob.Reader) error) error {
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
	if layer.MediaType == mediatype.OCI1Layer || layer.MediaType == mediatype.OCI1LayerZstd {
		return fmt.Errorf("only layers with gzip compression are supported (not %s)", layer.MediaType)
	}
	if layer.MediaType == mediatype.OCI1LayerGzip || layer.MediaType == mediatype.Docker2LayerGzip {

		d, err := digest.Parse(string(layer.Digest))
		if err != nil {
			return fmt.Errorf("failed to parse digest %s: %s", layer.Digest, err)
		}

		blob, err := rc.BlobGet(context.Background(), r, descriptor.Descriptor{Digest: d})
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
