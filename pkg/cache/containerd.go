package cache

import (
	"fmt"
	"io"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/uniget-org/cli/pkg/containers"
)

type ContainerdCache struct {
	namespace string
	client    *containerd.Client
}

func NewContainerdCache(namespace string) (*ContainerdCache, error) {
	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}
	//nolint:errcheck
	defer client.Close()

	return &ContainerdCache{
		namespace: namespace,
		client:    client,
	}, nil
}

func (c *ContainerdCache) Get(tool *containers.ToolRef, callback func(reader io.ReadCloser) error) error {
	err := containers.GetFirstLayerFromContainerdImage(c.client, tool, func(reader io.ReadCloser) error {
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
