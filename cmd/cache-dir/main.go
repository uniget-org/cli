package main

import (
	"os"
	"strings"

	"github.com/uniget-org/cli/pkg/archive"
	ucache "github.com/uniget-org/cli/pkg/cache"
)

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

	//cache := ucache.NewFileCache(cacheDirectory)
	cache := ucache.NewNoneCache()
	layer, err := cache.Get(ucache.NewToolRef(registry, repository, tool, version))
	if err != nil {
		panic(err)
	}

	err = archive.ProcessTarContents(layer, func(path string) string { return path } , func(path string) {})
	if err != nil {
		panic(err)
	}
}