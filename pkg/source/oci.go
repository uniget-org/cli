package source

import (
	"context"
	"fmt"
	"io"
	"strings"

	rref "github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

type OciBackend struct {
	Backend
}

func NewOciDownloader(cacheType cache.CacheType, cacheConfiguration CacheConfiguration) (*OciBackend, error) {
	backend, err := NewBackend(cacheType, cacheConfiguration)
	if err != nil {
		return nil, err
	}

	return &OciBackend{
		Backend: *backend,
	}, nil
}

func IsOciRef(source *Source) bool {
	return strings.HasPrefix(source.Url, "oci://")
}

func (d *OciBackend) Get(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	switch d.CacheType {
	case cache.CacheNone:
		return d.GetFromRegistry(source, p, callback)
	case cache.CacheFile:
		return d.GetFileCache(source, p, callback)
	case cache.CacheDocker:
		return d.HandleCache(source, p, callback)
	case cache.CacheContainerd:
		return d.HandleCache(source, p, callback)
	default:
		return fmt.Errorf("unsupported cache type: %v", d.CacheType)
	}
}

func (d *OciBackend) GetFromRegistry(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	ctx := context.Background()

	ref := strings.TrimPrefix(source.Url, "oci://")
	r, err := rref.New(ref)
	if err != nil {
		return fmt.Errorf("failed to create reference for %s: %w", ref, err)
	}

	rc := containers.GetRegclient()
	//nolint:errcheck
	defer rc.Close(ctx, r)

	logging.Debugf("NoneCache: Pulling %s", r)
	err = containers.GetFirstLayerFromRegistryRaw(ctx, rc, r, p, func(reader io.ReadCloser) error {
		err := callback(reader)
		if err != nil {
			return fmt.Errorf("failed to execute callback: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get layer for ref %s: %w", ref, err)
	}

	return nil
}

func (d *OciBackend) GetFileCache(source *Source, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	if !d.Cache.Has(source.Url) {
		err := d.GetFromRegistry(source, p, func(reader io.ReadCloser) error {
			return d.Cache.Put(source.Url, p, reader)
		})
		if err != nil {
			return err
		}
	}

	return d.Cache.Get(source.Url, p, callback)
}
