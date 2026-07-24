package source

import (
	"fmt"
	"io"
	"strconv"

	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type CacheConfiguration map[string]string

type Backend struct {
	Cache              cache.Cache
	CacheType          cache.CacheType
	CacheConfiguration CacheConfiguration
}

func NewBackend(cacheType cache.CacheType, cacheConfiguration CacheConfiguration) (*Backend, error) {
	backend := &Backend{
		CacheType:          cacheType,
		CacheConfiguration: cacheConfiguration,
	}
	err := backend.InitCache()
	if err != nil {
		return nil, err
	}

	return backend, nil
}

func (b *Backend) InitCache() (err error) {
	switch b.CacheType {
	case cache.CacheNone:
		b.Cache = cache.NewNoneCache()
	case cache.CacheFile:
		retentionSeconds, err := strconv.ParseInt(b.CacheConfiguration["retention_seconds"], 10, 0)
		if err != nil {
			return fmt.Errorf("failed to parse retention_seconds: %w", err)
		}
		b.Cache, err = cache.NewFileCache(b.CacheConfiguration["directory"], int(retentionSeconds))
		if err != nil {
			return fmt.Errorf("failed to initialize file cache: %w", err)
		}
	case cache.CacheDocker:
		b.Cache, err = cache.NewDockerCache()
		if err != nil {
			return fmt.Errorf("failed to initialize docker cache: %w", err)
		}
	case cache.CacheContainerd:
		b.Cache, err = cache.NewContainerdCache(b.CacheConfiguration["namespace"])
		if err != nil {
			return fmt.Errorf("failed to initialize containerd cache: %w", err)
		}
	default:
		return fmt.Errorf("unsupported cache type: %v", b.CacheType)
	}
	return nil
}

func NewBackendFromScheme(url *Source, cacheType cache.CacheType, cacheConfiguration CacheConfiguration) (Downloader, error) {
	if IsFileRef(url) {
		return NewFileDownloader(cacheType, cacheConfiguration)

	} else if IsWebRef(url) {
		return NewWebDownloader(cacheType, cacheConfiguration)

	} else if IsOciRef(url) {
		return NewOciDownloader(cacheType, cacheConfiguration)
	}

	return nil, fmt.Errorf("unsupported scheme for url: %s", url.Url)
}

func (b *Backend) HandleCache(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	if !b.Cache.Has(source.Url) {
		err := b.Cache.Put(source.Url, p, nil)
		if err != nil {
			return fmt.Errorf("failed to put %s into cache of type %s: %w", source.Url, b.Cache.GetName(), err)
		}
	}

	err := b.Cache.Get(source.Url, p, callback)
	if err != nil {
		return fmt.Errorf("failed to get %s from cache of type %s: %w", source.Url, b.Cache.GetName(), err)
	}

	return nil
}

type Source struct {
	Url string
}

type Downloader interface {
	Get(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error
}
