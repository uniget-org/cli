package main

import (
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/uniget-org/cli/pkg/containers"
)

func main() {
	registry := "ghcr.io"
	repository := "uniget-org/tools"
	tool := "jq"
	tag := "latest"
	ref := fmt.Sprintf("%s/%s/%s:%s", registry, repository, tool, tag)

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	layer, err := containers.GetFirstLayerFromDockerImage(cli, ref)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(fmt.Sprintf("%s-%s.tar", tool, tag), layer, 0644) // #nosec G306 -- just for testing
	if err != nil {
		panic(err)
	}
}