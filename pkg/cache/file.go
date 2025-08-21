package cache

import (
	"fmt"
	"io"
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

func (c *FileCache) writeDataToCache(reader io.ReadCloser, ref string) error {
	//nolint:errcheck
	defer reader.Close()

	if !c.cacheDirectoryExists() {
		return fmt.Errorf("cache directory is not set")
	}

	logging.Tracef("Writing data to cache for ref %s", ref)

	file, err := os.Create(fmt.Sprintf("%s/%s", c.cacheDirectory, ref))
	if err != nil {
		return fmt.Errorf("failed to create cache file for ref %s: %s", ref, err)
	}
	//nolint:errcheck
	defer file.Close()

	_, err = io.Copy(file, reader)
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
	if os.IsNotExist(err) {
		return false
	}

	expiredTime := stat.ModTime().Add(time.Duration(c.retentionSeconds) * time.Second)
	if expiredTime.Before(time.Now()) {
		logging.Debugf("Cache entry for ref %s expired", ref)
		return false
	}

	return true
}

func (c *FileCache) readDataFromCache(ref string, callback func(reader io.ReadCloser) error) error {
	if !c.cacheDirectoryExists() {
		return fmt.Errorf("cache directory is not set")
	}

	logging.Tracef("Reading data from cache for ref %s", ref)
	fileReader, err := os.Open(fmt.Sprintf("%s/%s", c.cacheDirectory, ref))
	if err != nil {
		return fmt.Errorf("failed to open cache file for ref %s: %s", ref, err)
	}
	//nolint:errcheck
	defer fileReader.Close()
	return callback(fileReader)
}

func (c *FileCache) Get(tool *containers.ToolRef, callback func(reader io.ReadCloser) error) error {
	cacheKey := tool.Key()
	if !c.checkDataInCache(tool.String()) {
		logging.Debugf("FileCache: Cache miss for %s", tool.String())
		err := c.n.Get(tool, func(reader io.ReadCloser) error {
			logging.Debugf("FileCache: Caching %s", tool.String())
			err := c.writeDataToCache(reader, cacheKey)
			if err != nil {
				return fmt.Errorf("failed to cache layer for ref %s: %w", tool, err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to get layer for ref %s: %w", tool, err)
		}

	}

	logging.Debugf("FileCache: Using cache for %s", tool.String())
	err := c.readDataFromCache(cacheKey, func(reader io.ReadCloser) error {
		logging.Debugf("FileCache: Reading cached data for %s", tool.String())
		return callback(reader)
	})
	if err != nil {
		return fmt.Errorf("failed to read layer for ref %s: %w", tool, err)
	}

	return nil
}
