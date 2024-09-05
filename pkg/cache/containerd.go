package cache

import (
	"fmt"

	"github.com/containerd/containerd"
	"github.com/uniget-org/cli/pkg/containers"
)

type ContainerdCache struct {
	registry string
	repository string
	toolSeparator string
	namespace string
	client *containerd.Client
}

func NewContainerdCache(namespace string, registry string, repository string, toolSeparator string) (*ContainerdCache, error) {
	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	return &ContainerdCache{
		registry: registry,
		repository: repository,
		toolSeparator: toolSeparator,
		namespace: namespace,
		client: client,
	}, nil
}

func (d *ContainerdCache) buildImageRef(key string) string {
	return fmt.Sprintf("%s/%s%s%s", d.registry, d.repository, d.toolSeparator, key)
}

func (d *ContainerdCache) WriteDataToCache(data []byte, key string) error {
	ref := d.buildImageRef(key)
	err := containers.PullContainerdImage(d.client, ref)
	if err != nil {
		return fmt.Errorf("failed to populate cache: %w", err)
	}

	return nil
}

func (d *ContainerdCache) CheckDataInCache(key string) bool {
	ref := d.buildImageRef(key)
	return containers.CheckContainerdImageExists(d.client, ref)
}

func (d *ContainerdCache) ReadDataFromCache(key string) ([]byte, error) {
	ref := d.buildImageRef(key)
	layer, err := containers.GetFirstLayerFromContainerdImage(d.client, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get layer from containerd image: %w", err)
	}
	return layer, nil
}