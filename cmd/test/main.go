package main

import (
	"fmt"

	"gitlab.com/uniget-org/cli/pkg/metadata"
	"gitlab.com/uniget-org/cli/pkg/source"
	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

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

	tools, err := unigetMetadataSource.Load()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded tools: %d\n", len(tools.Tools))
}
