package cache

import (
	"fmt"
	"io"

	containerd "github.com/containerd/containerd/v2/client"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type ContainerdCache struct {
	CacheStruct
	namespace string
	client    *containerd.Client
}

func NewContainerdCache(namespace string) (*ContainerdCache, error) {
	client, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	return &ContainerdCache{
		CacheStruct: CacheStruct{
			Type: CacheContainerd,
		},
		namespace: namespace,
		client:    client,
	}, nil
}

func (c *ContainerdCache) Put(key string, p tui.ProgressReader, reader io.ReadCloser) error {
	return containers.PullContainerdImage(c.client, key)
}

func (c *ContainerdCache) Has(key string) bool {
	return containers.CheckContainerdImageExists(c.client, key)
}

func (c *ContainerdCache) Get(key string, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	return containers.ReadContainerdImage(c.client, key, p, func(reader io.ReadCloser) error {
		err := callback(reader)
		if err != nil {
			return fmt.Errorf("failed to execute callback: %w", err)
		}
		return nil
	})
}
