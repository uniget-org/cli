package main

import (
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
)

func GetFirstLayerFromDockerImage(cli *client.Client, ref string) ([]byte, error) {
	shaString, err := containers.GetFirstLayerShaFromRegistry(ref)
	if err != nil {
		panic(err)
	}
	sha := shaString[7:]

	image, err := containers.GetDockerImage(cli, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %s", err)
	}

	layerGzip, err := containers.UnpackLayerFromDockerImage(image, sha)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack layer: %s", err)
	}

	layer, err := archive.Gunzip(layerGzip)
	if err != nil {
		return nil, fmt.Errorf("failed to gunzip layer: %s", err)
	}

	return layer, nil
}

func main() {
	registry := "ghcr.io"
	repository := "uniget-org/tools"
	tool := "jq"
	tag := "latest"
	ref := fmt.Sprintf("%s/%s/%s:%s", registry, repository, tool, tag)

	// TODO: Support equivalent of `docker --host=...`
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	layer, err := GetFirstLayerFromDockerImage(cli, ref)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(fmt.Sprintf("%s-%s.tar", tool, tag), layer, 0644) // #nosec G306 -- just for testing
	if err != nil {
		panic(err)
	}
}