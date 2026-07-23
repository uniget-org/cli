package cache

import (
	"io"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

type NoneCache struct {
	CacheStruct
}

func NewNoneCache() *NoneCache {
	return &NoneCache{
		CacheStruct: CacheStruct{
			Type: CacheNone,
		},
	}
}

func (c *NoneCache) Put(key string, p tui.ProgressReader, reader io.ReadCloser) error {
	return nil
}

func (c *NoneCache) Has(key string) bool {
	return false
}

func (c *NoneCache) Get(key string, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	return nil
}
