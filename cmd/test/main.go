package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory"
	"github.com/regclient/regclient/types/ref"
	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
)

// https://distribution.github.io/distribution/about/configuration/
const distributionConfig = `
version: 0.1
log:
  accesslog:
    disabled: true
  formatter: text
storage:
  inmemory:
  #filesystem:
  #  rootdirectory: /tmp/registry
`

var registryAddress = "127.0.0.1:5000"
var registryRepository = "uniget-org/tools"
var registryImage = "jq"
var registryTag = "1.7.1"

func startRegistry() {
	ctx := context.Background()

	config, err := configuration.Parse(bytes.NewReader([]byte(distributionConfig)))
	if err != nil {
		panic(err)
	}
	config.HTTP.Addr = registryAddress

	registry, err := registry.NewRegistry(ctx, config)
	if err != nil {
		panic(err)
	}
	err = registry.ListenAndServe()
	if err != nil {
		panic(err)
	}
	fmt.Println("DONE")
}

func addTestData() error {
	ctx := context.Background()
	rSrc, err := ref.New(fmt.Sprintf("%s/%s/%s:%s", "ghcr.io", registryRepository, registryImage, registryTag))
	if err != nil {
		return err
	}
	rTgt, err := ref.New(fmt.Sprintf("%s/%s/%s:%s", registryAddress, registryRepository, registryImage, registryTag))
	if err != nil {
		return err
	}

	rc := containers.GetRegclient()
	defer rc.Close(ctx, rSrc)
	defer rc.Close(ctx, rTgt)

	err = rc.ImageCopy(ctx, rSrc, rTgt)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	go startRegistry()

	err := addTestData()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	r := containers.NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	ref := r.GetRef()

	fmt.Println("Registry:")
	rc := containers.GetRegclient()
	defer rc.Close(ctx, ref)
	registryLayerGz, err := containers.GetFirstLayerFromRegistry(ctx, rc, ref)
	if err != nil {
		panic(err)
	}
	registryLayer, err := archive.Gunzip(registryLayerGz)
	if err != nil {
		panic(err)
	}
	err = archive.ProcessTarContents(registryLayer, func(reader *tar.Reader, header *tar.Header) error {
		//err := os.Chdir("/tmp")
		//if err != nil {
		//	return err
		//}

		return archive.CallbackExtractTarItem(reader, header)
	})
	if err != nil {
		panic(err)
	}

	os.Exit(0)

	fmt.Println("Docker:")
	cli, err := containers.GetDockerClient()
	if err != nil {
		panic(err)
	}
	dockerLayer, err := containers.GetFirstLayerFromDockerImage(cli, r)
	if err != nil {
		panic(err)
	}
	err = archive.ProcessTarContents(dockerLayer, archive.CallbackDisplayTarItem)
	if err != nil {
		panic(err)
	}

	os.Exit(0)

	fmt.Println("Containerd:")
	client, err := containers.GetContainerdClient()
	if err != nil {
		panic(err)
	}
	containerdLayer, err := containers.GetFirstLayerFromContainerdImage(client, r)
	if err != nil {
		panic(err)
	}
	err = archive.ProcessTarContents(containerdLayer, archive.CallbackDisplayTarItem)
	if err != nil {
		panic(err)
	}
}
