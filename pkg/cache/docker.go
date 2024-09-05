package cache

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/uniget-org/cli/pkg/containers"
)

type DockerCache struct {
	registry string
	repository string
	toolSeparator string
	cli *client.Client
}

func NewDockerCache(registry string, repository string, toolSeparator string) (*DockerCache, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerCache{
		registry: registry,
		repository: repository,
		toolSeparator: toolSeparator,
		cli: cli,
	}, nil
}

func (d *DockerCache) buildImageRef(key string) string {
	return fmt.Sprintf("%s/%s%s%s", d.registry, d.repository, d.toolSeparator, key)
}

func (d *DockerCache) WriteDataToCache(data []byte, key string) error {
	ref := d.buildImageRef(key)
	err := containers.PullDockerImage(d.cli, ref)
	if err != nil {
		return fmt.Errorf("failed to populate cache: %w", err)
	}

	return nil
}

func (d *DockerCache) CheckDataInCache(key string) bool {
	ref := d.buildImageRef(key)
	return containers.CheckDockerImageExists(d.cli, ref)
}

func (d *DockerCache) ReadDataFromCache(key string) ([]byte, error) {
	ref := d.buildImageRef(key)
	layer, err := containers.GetFirstLayerFromDockerImage(d.cli, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get layer from docker image: %w", err)
	}
	return layer, nil
}