package main

import (
	"fmt"

	"gitlab.com/uniget-org/cli/pkg/metadata"
	"gitlab.com/uniget-org/cli/pkg/source"
	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func main() {
	tarGzUnpacker := metadata.NewTarGzUnpacker()
	unigetMetadataSource, err := metadata.NewMetadataSource(
		"oci://ghcr.io/uniget-org/tools/metadata:main",
		".",
		cache.CacheNone,
		source.CacheConfiguration{},
		&tarGzUnpacker,
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

	nullUnpacker := metadata.NewNullUnpacker()
	var nullVerifier metadata.MetadataVerifier = metadata.NewNullMetadataVerifier()
	localMetadataSource, err := metadata.NewMetadataSource(
		"file:///home/nicholas/private/uniget/tools/metadata.json",
		".",
		cache.CacheNone,
		source.CacheConfiguration{},
		&nullUnpacker,
		map[string]string{
			"metadata.json":               "metadata.json",
			"metadata.json.sigstore.json": "metadata.json.sigstore.json",
		},
		&nullVerifier,
	)
	if err != nil {
		panic(err)
	}

	localTools, err := localMetadataSource.Load()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded local tools: %d\n", len(localTools.Tools))
}
