package cache

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/uniget-org/cli/pkg/containers"
)

type DockerCache struct {
	cli *client.Client
}

func NewDockerCache() (*DockerCache, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerCache{
		cli: cli,
	}, nil
}

func (c *DockerCache) Get(tool *ToolRef) ([]byte, error) {
	layer, err := containers.GetFirstLayerFromDockerImage(c.cli, tool.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get layer for ref %s: %w", tool, err)
	}

	return layer, nil
}