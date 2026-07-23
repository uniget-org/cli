package source

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

// fakeCache is a test double implementing cache.Cache.
type fakeCache struct {
	name    string
	has     bool
	putErr  error
	getErr  error
	putCall int
	getCall int
	hasCall int
	lastKey string
}

func (c *fakeCache) GetName() string { return c.name }

func (c *fakeCache) Put(key string, p tui.ProgressReader, reader io.ReadCloser) error {
	c.putCall++
	c.lastKey = key
	return c.putErr
}

func (c *fakeCache) Has(key string) bool {
	c.hasCall++
	c.lastKey = key
	return c.has
}

func (c *fakeCache) Get(key string, p tui.ProgressReader, callback func(reader io.ReadCloser) error) error {
	c.getCall++
	c.lastKey = key
	return c.getErr
}

func TestNewBackend(t *testing.T) {
	t.Run("none cache", func(t *testing.T) {
		b, err := NewBackend(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewBackend() unexpected error: %v", err)
		}
		if b.CacheType != cache.CacheNone { // #nosec SA5011 - Test code, b is not nil
			t.Errorf("CacheType = %v, want %v", b.CacheType, cache.CacheNone)
		}
		if b.Cache == nil || b.Cache.GetName() != "none" {
			t.Errorf("Cache = %v, want none cache", b.Cache)
		}
	})

	t.Run("file cache", func(t *testing.T) {
		b, err := NewBackend(cache.CacheFile, CacheConfiguration{
			"directory":         t.TempDir(),
			"retention_seconds": "60",
		})
		if err != nil {
			t.Fatalf("NewBackend() unexpected error: %v", err)
		}
		if b.CacheType != cache.CacheFile {
			t.Errorf("CacheType = %v, want %v", b.CacheType, cache.CacheFile)
		}
		if b.Cache == nil || b.Cache.GetName() != "file" {
			t.Errorf("Cache = %v, want file cache", b.Cache)
		}
	})

	t.Run("file cache invalid retention returns error", func(t *testing.T) {
		_, err := NewBackend(cache.CacheFile, CacheConfiguration{
			"directory":         t.TempDir(),
			"retention_seconds": "not-a-number",
		})
		if err == nil {
			t.Fatal("NewBackend() expected error for invalid retention_seconds, got nil")
		}
	})

	t.Run("unsupported cache type", func(t *testing.T) {
		_, err := NewBackend(cache.CacheType(999), nil)
		if err == nil {
			t.Fatal("NewBackend() expected error for unsupported cache type, got nil")
		}
	})
}

func TestBackend_InitCache(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		b := &Backend{CacheType: cache.CacheNone}
		if err := b.InitCache(); err != nil {
			t.Fatalf("InitCache() unexpected error: %v", err)
		}
		if b.Cache == nil {
			t.Fatal("expected Cache to be set")
		}
		if got, want := b.Cache.GetName(), "none"; got != want {
			t.Errorf("Cache.GetName() = %q, want %q", got, want)
		}
	})

	t.Run("file", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "cache")
		b := &Backend{
			CacheType: cache.CacheFile,
			CacheConfiguration: CacheConfiguration{
				"directory":         dir,
				"retention_seconds": "60",
			},
		}
		if err := b.InitCache(); err != nil {
			t.Fatalf("InitCache() unexpected error: %v", err)
		}
		if got, want := b.Cache.GetName(), "file"; got != want {
			t.Errorf("Cache.GetName() = %q, want %q", got, want)
		}
	})

	t.Run("file invalid retention", func(t *testing.T) {
		b := &Backend{
			CacheType: cache.CacheFile,
			CacheConfiguration: CacheConfiguration{
				"directory":         t.TempDir(),
				"retention_seconds": "not-a-number",
			},
		}
		if err := b.InitCache(); err == nil {
			t.Fatal("InitCache() expected error for invalid retention_seconds, got nil")
		}
	})

	t.Run("file mkdir failure", func(t *testing.T) {
		parent := t.TempDir()
		blocker := filepath.Join(parent, "file")
		if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
			t.Fatalf("setup: WriteFile: %v", err)
		}
		b := &Backend{
			CacheType: cache.CacheFile,
			CacheConfiguration: CacheConfiguration{
				"directory":         filepath.Join(blocker, "sub"),
				"retention_seconds": "60",
			},
		}
		if err := b.InitCache(); err == nil {
			t.Fatal("InitCache() expected error when directory cannot be created, got nil")
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		b := &Backend{CacheType: cache.CacheType(999)}
		if err := b.InitCache(); err == nil {
			t.Fatal("InitCache() expected error for unsupported cache type, got nil")
		}
	})
}

func TestBackend_HandleCache(t *testing.T) {
	src := &Source{Url: "some-url"}

	t.Run("miss populates then reads", func(t *testing.T) {
		fc := &fakeCache{name: "fake", has: false}
		b := &Backend{Cache: fc}

		err := b.HandleCache(src, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err != nil {
			t.Fatalf("HandleCache() unexpected error: %v", err)
		}
		if fc.putCall != 1 {
			t.Errorf("Put call count = %d, want 1", fc.putCall)
		}
		if fc.getCall != 1 {
			t.Errorf("Get call count = %d, want 1", fc.getCall)
		}
		if fc.lastKey != src.Url {
			t.Errorf("lastKey = %q, want %q", fc.lastKey, src.Url)
		}
	})

	t.Run("hit skips put", func(t *testing.T) {
		fc := &fakeCache{name: "fake", has: true}
		b := &Backend{Cache: fc}

		err := b.HandleCache(src, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err != nil {
			t.Fatalf("HandleCache() unexpected error: %v", err)
		}
		if fc.putCall != 0 {
			t.Errorf("Put call count = %d, want 0", fc.putCall)
		}
		if fc.getCall != 1 {
			t.Errorf("Get call count = %d, want 1", fc.getCall)
		}
	})

	t.Run("put error", func(t *testing.T) {
		fc := &fakeCache{name: "fake", has: false, putErr: errors.New("boom")}
		b := &Backend{Cache: fc}

		err := b.HandleCache(src, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err == nil {
			t.Fatal("HandleCache() expected error from Put, got nil")
		}
		if fc.getCall != 0 {
			t.Errorf("Get should not be called after Put error, got %d calls", fc.getCall)
		}
	})

	t.Run("get error", func(t *testing.T) {
		fc := &fakeCache{name: "fake", has: true, getErr: errors.New("boom")}
		b := &Backend{Cache: fc}

		err := b.HandleCache(src, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err == nil {
			t.Fatal("HandleCache() expected error from Get, got nil")
		}
	})
}
