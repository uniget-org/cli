package cache

import (
	"io"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

type CacheType int

const (
	CacheNone CacheType = iota
	CacheMemory
	CacheFile
	CacheDocker
	CacheContainerd
)

var CacheName = map[CacheType]string{
	CacheNone:       "none",
	CacheMemory:     "memory",
	CacheFile:       "file",
	CacheDocker:     "docker",
	CacheContainerd: "containerd",
}

type CacheStruct struct {
	Type CacheType
}

type Cache interface {
	GetName() string
	Put(key string, p tui.ProgressReader, reader io.ReadCloser) error
	Has(key string) bool
	Get(key string, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error
}

func (c *CacheStruct) GetName() string {
	return CacheName[c.Type]
}
