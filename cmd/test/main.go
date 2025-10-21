package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

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

	fmt.Println("Extracting image manifest tar to file foo")
	err = containers.ReadDockerImage(cli, image, func(reader io.ReadCloser) error {
		file, err := os.Create("foo")
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		//nolint:errcheck
		defer file.Close()
		_, err = io.Copy(file, reader)
		if err != nil {
			return fmt.Errorf("failed to copy data: %w", err)
		}

		return nil
	})
	if err != nil {
		panic(fmt.Errorf("failed to read docker image: %s", err))
	}

	fmt.Println("Extracting layer tar to file bar")
	err = containers.ReadDockerImage(cli, image, func(reader io.ReadCloser) error {
		err := containers.UnpackLayerFromDockerImage(reader, sha, func(reader io.ReadCloser) error {
			fmt.Println("  Unpacking...")

			file, err := os.Create("bar")
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			//nolint:errcheck
			defer file.Close()
			_, err = io.Copy(file, reader)
			if err != nil {
				return fmt.Errorf("failed to copy data: %w", err)
			}

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

	fmt.Println("Listing items in layer tar")
	err = containers.ReadDockerImage(cli, image, func(reader io.ReadCloser) error {
		err := containers.UnpackLayerFromDockerImage(reader, sha, func(reader io.ReadCloser) error {
			fmt.Println("  Unpacking...")

			buf := new(bytes.Buffer)
			length, err := buf.ReadFrom(reader)
			if err != nil {
				return fmt.Errorf("failed to copy data: %w", err)
			}
			fmt.Printf("  Read %d bytes\n", length)
			byteArray := buf.Bytes()
			fmt.Printf("  Total layer size: %d bytes\n", len(byteArray))

			fmt.Println("  Processing tar contents...")
			layerReader := bytes.NewReader(byteArray)
			err = archive.ProcessTarContents(io.NopCloser(layerReader), archive.CallbackDisplayTarItem)
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
