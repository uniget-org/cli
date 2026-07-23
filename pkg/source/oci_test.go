package source

import (
	"errors"
	"io"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func TestIsOciRef(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "oci scheme", url: "oci://example.com/foo:tag", want: true},
		{name: "http", url: "http://example.com/foo", want: false},
		{name: "file", url: "file:///tmp/foo", want: false},
		{name: "no scheme", url: "example.com/foo:tag", want: false},
		{name: "empty", url: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOciRef(&Source{Url: tt.url}); got != tt.want {
				t.Errorf("IsOciRef(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestNewOciDownloader(t *testing.T) {
	t.Run("none cache", func(t *testing.T) {
		d, err := NewOciDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewOciDownloader() unexpected error: %v", err)
		}
		if d.CacheType != cache.CacheNone {
			t.Errorf("CacheType = %v, want %v", d.CacheType, cache.CacheNone)
		}
		if d.Cache == nil || d.Cache.GetName() != "none" {
			t.Errorf("Cache = %v, want none cache", d.Cache)
		}
	})

	t.Run("file cache", func(t *testing.T) {
		d, err := NewOciDownloader(cache.CacheFile, CacheConfiguration{
			"directory":         t.TempDir(),
			"retention_seconds": "60",
		})
		if err != nil {
			t.Fatalf("NewOciDownloader() unexpected error: %v", err)
		}
		if d.Cache == nil || d.Cache.GetName() != "file" {
			t.Errorf("Cache = %v, want file cache", d.Cache)
		}
	})

	t.Run("invalid backend config propagates error", func(t *testing.T) {
		_, err := NewOciDownloader(cache.CacheFile, CacheConfiguration{
			"directory":         t.TempDir(),
			"retention_seconds": "not-a-number",
		})
		if err == nil {
			t.Fatal("NewOciDownloader() expected error for invalid retention_seconds, got nil")
		}
	})

	t.Run("unsupported cache type", func(t *testing.T) {
		_, err := NewOciDownloader(cache.CacheType(999), nil)
		if err == nil {
			t.Fatal("NewOciDownloader() expected error for unsupported cache type, got nil")
		}
	})
}

func TestOciBackend_Get(t *testing.T) {
	src := &Source{Url: "oci://example.com/foo:tag"}

	t.Run("unsupported cache type", func(t *testing.T) {
		d, err := NewOciDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewOciDownloader() unexpected error: %v", err)
		}
		d.CacheType = cache.CacheType(999)

		err = d.Get(src, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err == nil {
			t.Fatal("Get() expected error for unsupported cache type, got nil")
		}
	})

	t.Run("docker cache dispatches to HandleCache", func(t *testing.T) {
		d, err := NewOciDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewOciDownloader() unexpected error: %v", err)
		}
		d.CacheType = cache.CacheDocker
		fc := &fakeCache{name: "fake", has: true}
		d.Cache = fc

		err = d.Get(src, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if fc.getCall != 1 {
			t.Errorf("Cache.Get call count = %d, want 1", fc.getCall)
		}
		if fc.lastKey != src.Url {
			t.Errorf("Cache last key = %q, want %q", fc.lastKey, src.Url)
		}
	})

	t.Run("containerd cache dispatches to HandleCache", func(t *testing.T) {
		d, err := NewOciDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewOciDownloader() unexpected error: %v", err)
		}
		d.CacheType = cache.CacheContainerd
		fc := &fakeCache{name: "fake", has: true}
		d.Cache = fc

		err = d.Get(src, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if fc.getCall != 1 {
			t.Errorf("Cache.Get call count = %d, want 1", fc.getCall)
		}
	})
}

func TestOciBackend_GetFileCache_Hit(t *testing.T) {
	// When the cache already holds the ref we should never touch the registry.
	d, err := NewOciDownloader(cache.CacheFile, CacheConfiguration{
		"directory":         t.TempDir(),
		"retention_seconds": "60",
	})
	if err != nil {
		t.Fatalf("NewOciDownloader() unexpected error: %v", err)
	}
	fc := &fakeCache{name: "fake", has: true}
	d.Cache = fc

	src := &Source{Url: "oci://example.com/foo:tag"}
	err = d.GetFileCache(src, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
	if err != nil {
		t.Fatalf("GetFileCache() unexpected error: %v", err)
	}
	if fc.hasCall != 1 {
		t.Errorf("Cache.Has call count = %d, want 1", fc.hasCall)
	}
	if fc.getCall != 1 {
		t.Errorf("Cache.Get call count = %d, want 1", fc.getCall)
	}
	if fc.putCall != 0 {
		t.Errorf("Cache.Put should not be called on hit, got %d calls", fc.putCall)
	}
}

func TestOciBackend_GetFileCache_HitPropagatesGetError(t *testing.T) {
	d, err := NewOciDownloader(cache.CacheNone, nil)
	if err != nil {
		t.Fatalf("NewOciDownloader() unexpected error: %v", err)
	}
	want := errors.New("get failed")
	fc := &fakeCache{name: "fake", has: true, getErr: want}
	d.Cache = fc

	err = d.GetFileCache(&Source{Url: "oci://example.com/foo:tag"}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
	if !errors.Is(err, want) {
		t.Errorf("GetFileCache() error = %v, want %v", err, want)
	}
}

func TestOciBackend_GetFromRegistry_InvalidRef(t *testing.T) {
	d, err := NewOciDownloader(cache.CacheNone, nil)
	if err != nil {
		t.Fatalf("NewOciDownloader() unexpected error: %v", err)
	}

	// An uppercase letter in the repository segment is rejected by regclient.
	err = d.GetFromRegistry(&Source{Url: "oci://example.com/BAD:tag"}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
	if err == nil {
		t.Fatal("GetFromRegistry() expected error for invalid ref, got nil")
	}
}
