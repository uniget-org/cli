package main

import (
	"archive/tar"
	"context"
	"fmt"
	"os"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
)

var (
	registryAddress    = "127.0.0.1:5000"
	registryRepository = "uniget-org/tools"
	registryImage      = "jq"
	registryTag        = "1.7.1"
	r                  = containers.NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
)

func main() {
	ctx := context.Background()

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
