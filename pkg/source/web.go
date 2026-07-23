package source

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type WebBackend struct {
	Backend
}

func NewWebDownloader(cacheType cache.CacheType, cacheConfiguration CacheConfiguration) (*WebBackend, error) {
	if cacheType != cache.CacheNone && cacheType != cache.CacheFile {
		return nil, fmt.Errorf("web backend only supports caching to file")
	}

	backend, err := NewBackend(cacheType, cacheConfiguration)
	if err != nil {
		return nil, err
	}

	return &WebBackend{
		Backend: *backend,
	}, nil
}

func IsWebRef(source *Source) bool {
	return strings.HasPrefix(source.Url, "http://") || strings.HasPrefix(source.Url, "https://")
}

func (d *WebBackend) Get(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) (err error) {
	switch d.CacheType {
	case cache.CacheNone:
		resp, err := http.Get(source.Url)
		if err != nil {
			return fmt.Errorf("failed to get %s: %w", source.Url, err)
		}
		//nolint:errcheck
		defer resp.Body.Close()

		return callback(resp.Body)

	case cache.CacheFile:
		if !d.Cache.Has(source.Url) {
			resp, err := http.Get(source.Url)
			if err != nil {
				return fmt.Errorf("failed to get %s: %w", source.Url, err)
			}
			//nolint:errcheck
			defer resp.Body.Close()

			err = d.Cache.Put(source.Url, p, resp.Body)
			if err != nil {
				return fmt.Errorf("failed to put %s into cache: %w", source.Url, err)
			}
		}
	}

	return d.Cache.Get(source.Url, p, callback)
}
