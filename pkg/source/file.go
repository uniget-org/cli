package source

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type FileBackend struct {
	Backend
}

func NewFileDownloader(cacheType cache.CacheType, cacheConfiguration CacheConfiguration) (*FileBackend, error) {
	if cacheType != cache.CacheNone {
		return nil, fmt.Errorf("file backend does not support caching")
	}

	backend, err := NewBackend(cacheType, cacheConfiguration)
	if err != nil {
		return nil, err
	}

	return &FileBackend{
		Backend: *backend,
	}, nil
}

func IsFileRef(source *Source) bool {
	return strings.HasPrefix(source.Url, "file://")
}

func (d *FileBackend) Get(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) (err error) {
	f, err := os.Open(strings.TrimPrefix(source.Url, "file://"))
	if err != nil {
		return err
	}
	//nolint:errcheck
	defer f.Close()

	return callback(f)
}
