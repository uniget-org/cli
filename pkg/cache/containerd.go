package cache

import (
	"fmt"

	"github.com/containerd/containerd"
	"github.com/uniget-org/cli/pkg/containers"
)

type ContainerdCache struct {
	namespace string
	client *containerd.Client
}

func NewContainerdCache(namespace string, ) (*ContainerdCache, error) {
	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	return &ContainerdCache{
		namespace: namespace,
		client: client,
	}, nil
}

func (c *ContainerdCache) Get(tool *ToolRef) ([]byte, error) {
	layer, err := containers.GetFirstLayerFromContainerdImage(c.client, tool.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get layer for ref %s: %w", tool, err)
	}

	return layer, nil
}