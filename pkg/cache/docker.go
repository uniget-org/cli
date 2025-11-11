package cache

import (
	"fmt"
	"io"

	"github.com/moby/moby/client"
	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/logging"
)

type DockerCache struct {
	cli *client.Client
}

func NewDockerCache() (*DockerCache, error) {
	logging.Tracef("Creating Docker cache")
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerCache{
		cli: cli,
	}, nil
}

func (c *DockerCache) Get(tool *containers.ToolRef, callback func(reader io.ReadCloser) error) error {
	logging.Debugf("DockerCache: Pulling %s", tool)
	err := containers.GetFirstLayerFromDockerImage(c.cli, tool, func(reader io.ReadCloser) error {
		err := callback(reader)
		if err != nil {
			return fmt.Errorf("failed to execute callback: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get layer for ref %s: %w", tool, err)
	}

	return nil
}
