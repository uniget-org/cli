package cache

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/logging"
)

type DockerCache struct {
	cli *client.Client
}

func NewDockerCache() (*DockerCache, error) {
	logging.Tracef("Creating Docker cache")
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerCache{
		cli: cli,
	}, nil
}

func (c *DockerCache) Get(tool *containers.ToolRef) ([]byte, error) {
	logging.Debugf("DockerCache: Pulling %s", tool)
	layer, err := containers.GetFirstLayerFromDockerImage(c.cli, tool)
	if err != nil {
		return nil, fmt.Errorf("failed to get layer for ref %s: %w", tool, err)
	}

	return layer, nil
}
