package cache

import (
	"fmt"
	"os"

	"github.com/uniget-org/cli/pkg/containers"
)

type FileCache struct {
	n              *NoneCache
	cacheDirectory string
}

func NewFileCache(directory string) *FileCache {
	return &FileCache{
		n:              NewNoneCache(),
		cacheDirectory: directory,
	}
}

func (c *FileCache) cacheDirectoryExists() bool {
	if c.cacheDirectory == "" {
		return false
	}

	_, err := os.Stat(c.cacheDirectory)
	return !os.IsNotExist(err)
}

func (c *FileCache) writeDataToCache(data []byte, ref string) error {
	if !c.cacheDirectoryExists() {
		return fmt.Errorf("cache directory is not set")
	}

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

	_, err := os.Stat(fmt.Sprintf("%s/%s", c.cacheDirectory, ref))
	return !os.IsNotExist(err)
}

func (c *FileCache) readDataFromCache(ref string) ([]byte, error) {
	if !c.cacheDirectoryExists() {
		return nil, fmt.Errorf("cache directory is not set")
	}

	data, err := os.ReadFile(fmt.Sprintf("%s/%s", c.cacheDirectory, ref))
	if err != nil {
		return nil, fmt.Errorf("failed to read data for ref %s from cache: %s", ref, err)
	}
	return data, nil
}

func (c *FileCache) Get(tool *containers.ToolRef) ([]byte, error) {
	cacheKey := tool.Key()
	if !c.checkDataInCache(tool.String()) {
		layer, err := c.n.Get(tool)
		if err != nil {
			panic(err)
		}

		err = c.writeDataToCache(layer, cacheKey)
		if err != nil {
			panic(err)
		}
	}

	layer, err := c.readDataFromCache(cacheKey)
	if err != nil {
		panic(err)
	}

	return layer, nil
}
