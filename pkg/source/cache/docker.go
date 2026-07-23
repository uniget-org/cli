package cache

import (
	"fmt"
	"io"

	"github.com/moby/moby/client"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type DockerCache struct {
	CacheStruct
	cli *client.Client
}

func NewDockerCache() (*DockerCache, error) {
	logging.Tracef("Creating Docker cache")
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerCache{
		CacheStruct: CacheStruct{
			Type: CacheDocker,
		},
		cli: cli,
	}, nil
}

func (c *DockerCache) Put(key string, p tui.ProgressReader, reader io.ReadCloser) error {
	err := containers.PullDockerImage(c.cli, key)
	if err != nil {
		return fmt.Errorf("failed to pull docker image: %w", err)
	}

	return nil
}

func (c *DockerCache) Has(key string) bool {
	return containers.CheckDockerImageExists(c.cli, key)
}

func (c *DockerCache) Get(key string, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	return containers.ReadDockerImage(c.cli, key, p, func(reader io.ReadCloser) error {
		err := callback(reader)
		if err != nil {
			return fmt.Errorf("failed to execute callback: %w", err)
		}
		return nil
	})
}
