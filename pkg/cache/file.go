package cache

import (
	"fmt"
	"os"
	"time"

	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/logging"
)

type FileCache struct {
	n                *NoneCache
	cacheDirectory   string
	retentionSeconds int
}

func NewFileCache(directory string, retentionSeconds int) *FileCache {
	return &FileCache{
		n:                NewNoneCache(),
		cacheDirectory:   directory,
		retentionSeconds: retentionSeconds,
	}
}

func (c *FileCache) cacheDirectoryExists() bool {
	if c.cacheDirectory == "" {
		return false
	}

	logging.Tracef("Checking cache directory %s", c.cacheDirectory)
	_, err := os.Stat(c.cacheDirectory)
	return !os.IsNotExist(err)
}

func (c *FileCache) writeDataToCache(data []byte, ref string) error {
	if !c.cacheDirectoryExists() {
		return fmt.Errorf("cache directory is not set")
	}

	logging.Tracef("Writing data to cache for ref %s", ref)
	err := os.WriteFile(fmt.Sprintf("%s/%s", c.cacheDirectory, ref), data, 0644) // #nosec G306 -- just for testing
	if err != nil {
		return fmt.Errorf("failed to write data for ref %s to cache: %s", ref, err)
	}
	return nil
}

func (c *FileCache) checkDataInCache(ref string) bool {
	if !c.cacheDirectoryExists() {
		return false
	}

	logging.Tracef("Checking cache for ref %s", ref)
	stat, err := os.Stat(fmt.Sprintf("%s/%s", c.cacheDirectory, ref))
	if !os.IsNotExist(err) {
		return false
	}

	expiredTime := stat.ModTime().Add(time.Duration(c.retentionSeconds) * time.Second)
	if expiredTime.Before(time.Now()) {
		logging.Debugf("Cache entry for ref %s expired", ref)
		return false
	}

	return true
}

func (c *FileCache) readDataFromCache(ref string) ([]byte, error) {
	if !c.cacheDirectoryExists() {
		return nil, fmt.Errorf("cache directory is not set")
	}

	logging.Tracef("Reading data from cache for ref %s", ref)
	data, err := os.ReadFile(fmt.Sprintf("%s/%s", c.cacheDirectory, ref))
	if err != nil {
		return nil, fmt.Errorf("failed to read data for ref %s from cache: %s", ref, err)
	}
	return data, nil
}

func (c *FileCache) Get(tool *containers.ToolRef) ([]byte, error) {
	cacheKey := tool.Key()
	if !c.checkDataInCache(tool.String()) {
		logging.Debugf("FileCache: Cache miss for %s", tool.String())
		layer, err := c.n.Get(tool)
		if err != nil {
			panic(err)
		}

		logging.Debugf("FileCache: Caching %s", tool.String())
		err = c.writeDataToCache(layer, cacheKey)
		if err != nil {
			panic(err)
		}
	}

	logging.Debugf("FileCache: Using cache for %s", tool.String())
	layer, err := c.readDataFromCache(cacheKey)
	if err != nil {
		panic(err)
	}

	return layer, nil
}
