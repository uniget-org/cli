package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/ref"
	"github.com/uniget-org/cli/pkg/cache"
	"github.com/uniget-org/cli/pkg/containers"
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
	toolRef := fmt.Sprintf("%s/%s/%s:%s", registry, repository, tool, version)

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

	var cache = cache.NewFileCache(cacheDirectory)

	cacheKey := fmt.Sprintf("%s-%s", tool, version)
	if cache.CheckDataInCache(cacheKey) {
		fmt.Printf("Cache hit for %s\n", cacheKey)

	} else {
		fmt.Printf("Cache miss for %s\n", cacheKey)

		ctx := context.Background()

		r, err := ref.New(toolRef)
		if err != nil {
			panic(err)
		}

		rcOpts := []regclient.Opt{}
		rcOpts = append(rcOpts, regclient.WithUserAgent("uniget"))
		rcOpts = append(rcOpts, regclient.WithDockerCreds())
		rc := regclient.New(rcOpts...)
		defer rc.Close(ctx, r)

		layer, err := containers.GetFirstLayerFromRegistry(ctx, rc, r)
		if err != nil {
			panic(err)
		}

		err = cache.WriteDataToCache(layer, cacheKey)
		if err != nil {
			panic(err)
		}
	}

	layer, err := cache.ReadDataFromCache(cacheKey)
	if err != nil {
		panic(err)
	}

	err = containers.ProcessLayerContents(layer, func(path string) string { return path } , func(path string) {})
	if err != nil {
		panic(err)
	}
}