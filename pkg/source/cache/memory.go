package cache

import (
	"bytes"
	"io"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

type MemoryCache struct {
	CacheStruct
	data map[string][]byte
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		CacheStruct: CacheStruct{
			Type: CacheMemory,
		},
		data: make(map[string][]byte),
	}
}

func (c *MemoryCache) Put(key string, p tui.ProgressReader, reader io.ReadCloser) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	c.data[key] = data
	return nil
}

func (c *MemoryCache) Has(key string) bool {
	_, ok := c.data[key]
	return ok
}

func (c *MemoryCache) Get(key string, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	data, ok := c.data[key]
	if !ok {
		return nil
	}
	return callback(io.NopCloser(bytes.NewReader(data)))
}
