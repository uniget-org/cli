package main

import (
	"gitlab.com/uniget-org/cli/pkg/metadata"
	"gitlab.com/uniget-org/cli/pkg/source"
	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

var pathPrefix = "./"
var ociRef = "ghcr.io/uniget-org/tools/uniget:latest"
var webUri = "https://gitlab.com/uniget-org/cli/-/releases/permalink/latest/downloads/uniget_Linux_x86_64.tar.gz"

func main() {
	unigetMetadataSource, err := metadata.NewMetadataSource(
		"oci://ghcr.io/uniget-org/tools/metadata:main",
		".",
		cache.CacheNone,
		source.CacheConfiguration{},
		metadata.NewTarGzUnpacker(),
		map[string]string{
			"metadata.json":               "metadata.json",
			"metadata.json.sigstore.json": "metadata.json.sigstore.json",
		},
		metadata.NewSigstoreMetadataVerifier(
			"https://token.actions.githubusercontent.com",
			"",
			"",
			"https://github\\.com/uniget-org/tools/\\.github/workflows/[^.]+\\.yml@refs/heads/main",
		),
	)
	if err != nil {
		panic(err)
	}
	err = unigetMetadataSource.Download(tui.NewProgressReader(nil, nil))
	if err != nil {
		panic(err)
	}
}
