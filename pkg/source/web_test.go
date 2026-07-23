package source

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func TestIsWebRef(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "http", url: "http://example.com/foo", want: true},
		{name: "https", url: "https://example.com/foo", want: true},
		{name: "file", url: "file:///tmp/foo", want: false},
		{name: "oci", url: "oci://example.com/foo:tag", want: false},
		{name: "empty", url: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWebRef(&Source{Url: tt.url}); got != tt.want {
				t.Errorf("IsWebRef(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestNewWebDownloader(t *testing.T) {
	t.Run("none cache", func(t *testing.T) {
		d, err := NewWebDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewWebDownloader() unexpected error: %v", err)
		}
		if d.CacheType != cache.CacheNone { // #nosec SA5011 - Test code, d is not nil
			t.Errorf("CacheType = %v, want %v", d.CacheType, cache.CacheNone)
		}
		if d.Cache == nil || d.Cache.GetName() != "none" {
			t.Errorf("Cache = %v, want none cache", d.Cache)
		}
	})

	t.Run("file cache", func(t *testing.T) {
		d, err := NewWebDownloader(cache.CacheFile, CacheConfiguration{
			"directory":         t.TempDir(),
			"retention_seconds": "60",
		})
		if err != nil {
			t.Fatalf("NewWebDownloader() unexpected error: %v", err)
		}
		if d.CacheType != cache.CacheFile {
			t.Errorf("CacheType = %v, want %v", d.CacheType, cache.CacheFile)
		}
		if d.Cache == nil || d.Cache.GetName() != "file" {
			t.Errorf("Cache = %v, want file cache", d.Cache)
		}
	})

	t.Run("unsupported cache rejected", func(t *testing.T) {
		for _, ct := range []cache.CacheType{cache.CacheDocker, cache.CacheContainerd, cache.CacheMemory} {
			_, err := NewWebDownloader(ct, nil)
			if err == nil {
				t.Errorf("NewWebDownloader() with cache %v: expected error, got nil", ct)
			}
		}
	})
}

func TestWebBackend_Get(t *testing.T) {
	t.Run("cache miss triggers http then get", func(t *testing.T) {
		var hits int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&hits, 1)
			_, _ = w.Write([]byte("payload"))
		}))
		defer srv.Close()

		d, err := NewWebDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewWebDownloader() unexpected error: %v", err)
		}
		fc := &fakeCache{name: "fake", has: false}
		d.Cache = fc

		err = d.Get(&Source{Url: srv.URL}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if got := atomic.LoadInt32(&hits); got != 1 {
			t.Errorf("http server hits = %d, want 1", got)
		}
		if fc.putCall != 1 {
			t.Errorf("Cache.Put call count = %d, want 1", fc.putCall)
		}
		if fc.getCall != 1 {
			t.Errorf("Cache.Get call count = %d, want 1", fc.getCall)
		}
		if fc.lastKey != srv.URL {
			t.Errorf("Cache last key = %q, want %q", fc.lastKey, srv.URL)
		}
	})

	t.Run("cache hit skips http", func(t *testing.T) {
		var hits int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&hits, 1)
			_, _ = w.Write([]byte("payload"))
		}))
		defer srv.Close()

		d, err := NewWebDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewWebDownloader() unexpected error: %v", err)
		}
		fc := &fakeCache{name: "fake", has: true}
		d.Cache = fc

		err = d.Get(&Source{Url: srv.URL}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if got := atomic.LoadInt32(&hits); got != 0 {
			t.Errorf("http server hits = %d, want 0", got)
		}
		if fc.putCall != 0 {
			t.Errorf("Cache.Put call count = %d, want 0", fc.putCall)
		}
		if fc.getCall != 1 {
			t.Errorf("Cache.Get call count = %d, want 1", fc.getCall)
		}
	})

	t.Run("http error", func(t *testing.T) {
		// Reserve then close a listener to obtain a definitely-unused address.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		badURL := srv.URL
		srv.Close()

		d, err := NewWebDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewWebDownloader() unexpected error: %v", err)
		}
		fc := &fakeCache{name: "fake", has: false}
		d.Cache = fc

		err = d.Get(&Source{Url: badURL}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err == nil {
			t.Fatal("Get() expected error for unreachable server, got nil")
		}
		if fc.putCall != 0 {
			t.Errorf("Cache.Put should not be called on http failure, got %d calls", fc.putCall)
		}
	})

	t.Run("propagates cache get error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("payload"))
		}))
		defer srv.Close()

		d, err := NewWebDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewWebDownloader() unexpected error: %v", err)
		}
		want := errors.New("get failed")
		fc := &fakeCache{name: "fake", has: true, getErr: want}
		d.Cache = fc

		err = d.Get(&Source{Url: srv.URL}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if !errors.Is(err, want) {
			t.Errorf("Get() error = %v, want %v", err, want)
		}
	})

	t.Run("response body is piped into cache.Put", func(t *testing.T) {
		payload := []byte("body-bytes")
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(payload)
		}))
		defer srv.Close()

		d, err := NewWebDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewWebDownloader() unexpected error: %v", err)
		}
		rc := &recordingCache{}
		d.Cache = rc

		err = d.Get(&Source{Url: srv.URL}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if !bytes.Equal(rc.putBody, payload) {
			t.Errorf("Cache.Put received %q, want %q", rc.putBody, payload)
		}
	})
}

// recordingCache captures the body handed to Put for assertion.
type recordingCache struct {
	fakeCache
	putBody []byte
}

func (c *recordingCache) Put(key string, p tui.ProgressReader, reader io.ReadCloser) error {
	if reader != nil {
		b, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		c.putBody = b
	}
	c.putCall++
	c.lastKey = key
	return c.putErr
}
