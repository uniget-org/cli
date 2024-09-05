package main

import (
	"fmt"
	"os"

	ucache "github.com/uniget-org/cli/pkg/cache"
)

func main() {
	registry := "ghcr.io"
	repository := "uniget-org/tools"
	tool := "jq"
	tag := "latest"
	
	cache, err := ucache.NewDockerCache()
	if err != nil {
		panic(err)
	}

	layer, err := cache.Get(ucache.NewToolRef(registry, repository, tool, tag))
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(fmt.Sprintf("%s-%s.tar", tool, tag), layer, 0644) // #nosec G306 -- just for testing
	if err != nil {
		panic(err)
	}
}