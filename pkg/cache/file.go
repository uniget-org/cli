package cache

import (
	"fmt"
	"os"
)

var cacheDirectory string

type FileCache struct {
	cacheDirectory string
}

func NewFileCache(directory string) *FileCache {
	return &FileCache{
		cacheDirectory: directory,
	}
}

func (c *FileCache) checkIfCacheDirectoryExists() bool {
	if c.cacheDirectory == "" {
		return false
	}

	_, err := os.Stat(cacheDirectory)
	return !os.IsNotExist(err)
}

func (c *FileCache) WriteDataToCache(data []byte, key string) error {
	if ! c.checkIfCacheDirectoryExists() {
		return fmt.Errorf("cache directory is not set")
	}

	err := os.WriteFile(fmt.Sprintf("%s/%s", cacheDirectory, key), data, 0644) // #nosec G306 -- just for testing
	if err != nil {
		return fmt.Errorf("failed to write data for key %s to cache: %s", key, err)
	}
	return nil
}

func (c *FileCache) CheckDataInCache(key string) bool {
	if ! c.checkIfCacheDirectoryExists() {
		return false
	}

	_, err := os.Stat(fmt.Sprintf("%s/%s", cacheDirectory, key))
	return !os.IsNotExist(err)
}

func (c *FileCache) ReadDataFromCache(key string) ([]byte, error) {
	if ! c.checkIfCacheDirectoryExists() {
		return nil, fmt.Errorf("cache directory is not set")
	}

	data, err := os.ReadFile(fmt.Sprintf("%s/%s", cacheDirectory, key))
	if err != nil {
		return nil, fmt.Errorf("failed to read data for key %s from cache: %s", key, err)
	}
	return data, nil
}