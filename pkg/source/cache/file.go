package cache

import (
	"fmt"
	"io"
	"os"
	"time"

	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type FileCache struct {
	CacheStruct
	cacheDirectory   string
	retentionSeconds int
}

func NewFileCache(directory string, retentionSeconds int) (*FileCache, error) {
	cache := &FileCache{
		CacheStruct: CacheStruct{
			Type: CacheFile,
		},
		cacheDirectory:   directory,
		retentionSeconds: retentionSeconds,
	}

	// TODO: Sanity checks
	err := os.MkdirAll(directory, 0o750)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cache, nil
}

func (c *FileCache) Put(key string, p tui.ProgressReader, reader io.ReadCloser) error {
	//nolint:errcheck
	defer reader.Close()

	logging.Tracef("Writing data to cache for key %s", key)

	file, err := os.Create(fmt.Sprintf("%s/%s", c.cacheDirectory, key)) // #nosec G703 -- Base directory is meant to be configurable
	if err != nil {
		return fmt.Errorf("failed to create cache file for key %s: %s", key, err)
	}
	//nolint:errcheck
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to write data for key %s to cache: %s", key, err)
	}

	return nil
}

func (c *FileCache) Has(key string) bool {
	logging.Tracef("Checking cache for key %s", key)
	stat, err := os.Stat(fmt.Sprintf("%s/%s", c.cacheDirectory, key))
	if os.IsNotExist(err) {
		return false
	}

	expiredTime := stat.ModTime().Add(time.Duration(c.retentionSeconds) * time.Second)
	if expiredTime.Before(time.Now()) {
		logging.Debugf("Cache entry for key %s expired", key)

		// TODO: Sanity checks
		err = os.Remove(fmt.Sprintf("%s/%s", c.cacheDirectory, key))
		if err != nil {
			logging.Warning.Printfln("Failed to remove key %s from cache: %s", key, err)
		}
		return false
	}

	return true
}

func (c *FileCache) Get(key string, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	logging.Tracef("Reading data from cache for key %s", key)
	fileReader, err := os.Open(fmt.Sprintf("%s/%s", c.cacheDirectory, key))
	if err != nil {
		return fmt.Errorf("failed to open cache file for key %s: %s", key, err)
	}
	//nolint:errcheck
	defer fileReader.Close()
	return callback(fileReader)
}
