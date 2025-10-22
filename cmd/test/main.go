package main

import (
	"fmt"
	"io"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
)

func main() {
	toolRef := containers.NewToolRef("ghcr.io", "uniget-org/tools", "continue", "main")
	//ref := toolRef.GetRef()
	image := toolRef.String()

	cli, err := containers.GetDockerClient()
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	fmt.Printf("Image %s\n", image)
	if !containers.CheckDockerImageExists(cli, image) {
		fmt.Printf("Pulling %s\n", image)
		err := containers.PullDockerImage(cli, image)
		if err != nil {
			panic(fmt.Errorf("failed to pull docker image: %s", err))
		}
	}

	shaString, err := containers.GetFirstLayerShaFromRegistry(toolRef)
	if err != nil {
		panic(fmt.Errorf("failed to get first layer sha: %s", err))
	}
	sha := shaString[7:]
	fmt.Printf("Layer Sha256: %s\n", sha)

	fmt.Println("Listing items in layer tar")
	err = containers.ReadDockerImage(cli, image, func(reader io.ReadCloser) error {
		err := containers.UnpackLayerFromDockerImage(reader, sha, func(reader io.ReadCloser) error {
			fmt.Println("  Unpacking...")

			fmt.Println("  Processing tar contents...")
			err = archive.ProcessTarContents(io.NopCloser(reader), archive.CallbackDisplayTarItem)
			if err != nil {
				return fmt.Errorf("failed to process tar: %w", err)
			}
			fmt.Println("  Done processing tar contents.")

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to unpack layer: %w", err)
		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("failed to read docker image: %s", err))
	}

}
