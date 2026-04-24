package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"

	"github.com/google/safearchive/tar"
	"github.com/opencontainers/go-digest"
	"github.com/pterm/pterm"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"

	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/containers"
)

func GetFirstLayerFromManifest(ctx context.Context, rc *regclient.RegClient, m manifest.Manifest, p ProgressReader, callback func(reader io.ReadCloser) error) error {
	return GetLayerFromManifestByIndex(ctx, rc, m, 0, p, callback)
}

func GetLayerFromManifestByIndex(ctx context.Context, rc *regclient.RegClient, m manifest.Manifest, index int, p ProgressReader, callback func(reader io.ReadCloser) error) error {
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

	if len(layers) < index {
		return fmt.Errorf("image only has %d layers", len(layers))
	}

	layer := layers[index]
	p.SetTotal(layer.Size)

	if layer.MediaType == mediatype.OCI1Layer || layer.MediaType == mediatype.OCI1LayerZstd {
		return fmt.Errorf("only layers with gzip compression are supported (not %s)", layer.MediaType)
	}
	if layer.MediaType == mediatype.OCI1LayerGzip || layer.MediaType == mediatype.Docker2LayerGzip {

		d, err := digest.Parse(string(layer.Digest))
		if err != nil {
			return fmt.Errorf("failed to parse digest %s: %s", layer.Digest, err)
		}

		blob, err := rc.BlobGet(context.Background(), m.GetRef(), descriptor.Descriptor{Digest: d})
		if err != nil {
			return fmt.Errorf("failed to get blob for digest %s: %s", layer.Digest, err)
		}

		p.SetReader(blob)
		err = callback(p)
		if err != nil {
			return fmt.Errorf("failed to execute callback: %w", err)
		}

		return nil
	}

	return fmt.Errorf("unsupported layer media type %s", layer.MediaType)
}

func main() {
	toolRef := containers.NewToolRef("ghcr.io", "uniget-org/tools", "uniget", "0.26.4")
	ref := toolRef.GetRef()
	fmt.Printf("ref: %+v\n", ref)

	ctx := context.Background()
	rc := containers.GetRegclient()

	m, err := containers.GetPlatformManifestForLocalPlatform(ctx, rc, ref)
	if err != nil {
		panic(err)
	}

	progressPrinter, err := pterm.DefaultProgressbar.WithTitle("Downloading stuff").WithTotal(0).Start()
	if err != nil {
		panic(err)
	}
	p := NewProgressReader(
		func(n int64) {
			progressPrinter.Total = int(n)
		},
		func(n int64) {
			progressPrinter.Add(int(n))
		},
	)

	err = GetFirstLayerFromManifest(ctx, rc, m, p, func(reader io.ReadCloser) error {
		//fmt.Println("CallbackLayer()")

		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %s", err)
		}

		return archive.ProcessTarContents(gzipReader, func(tarReader *tar.Reader, header *tar.Header) error {
			if header.Typeflag == tar.TypeReg {
				//fmt.Printf("Processing tar item: %s\n", header.Name)
			}

			return nil
		})
	})
	if err != nil {
		panic(err)
	}
}
