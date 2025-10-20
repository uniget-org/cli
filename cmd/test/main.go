package main

import (
	"fmt"
	"io"

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
		containers.PullDockerImage(cli, image)
	}

	shaString, err := containers.GetFirstLayerShaFromRegistry(toolRef)
	if err != nil {
		panic(fmt.Errorf("failed to get first layer sha: %s", err))
	}
	sha := shaString[7:]
	fmt.Printf("Layer Sha256: %s\n", sha)

	err = containers.ReadDockerImage(cli, image, func(reader io.ReadCloser) error {
		fmt.Println("Reading image")
		defer reader.Close()

		fmt.Println("X")
		err := containers.UnpackLayerFromDockerImage(reader, sha, func(reader io.ReadCloser) error {
			fmt.Println("  Unpacking...")
			//
			fmt.Println("  Dome unpacking.")
			return nil
		})
		if err != nil {
			return err
		}
		fmt.Println("Y")

		fmt.Println("  Done reading.")
		return nil
	})

}
