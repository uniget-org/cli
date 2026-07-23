package source

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/source/cache"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func TestIsFileRef(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "file scheme", url: "file:///tmp/foo", want: true},
		{name: "http", url: "http://example.com/foo", want: false},
		{name: "oci", url: "oci://example.com/foo:tag", want: false},
		{name: "plain path", url: "/tmp/foo", want: false},
		{name: "empty", url: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFileRef(&Source{Url: tt.url}); got != tt.want {
				t.Errorf("IsFileRef(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestNewFileDownloader(t *testing.T) {
	t.Run("none cache", func(t *testing.T) {
		d, err := NewFileDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewFileDownloader() unexpected error: %v", err)
		}
		if d.CacheType != cache.CacheNone {
			t.Errorf("CacheType = %v, want %v", d.CacheType, cache.CacheNone)
		}
		if d.Cache == nil || d.Cache.GetName() != "none" {
			t.Errorf("Cache = %v, want none cache", d.Cache)
		}
	})

	t.Run("non-none cache rejected", func(t *testing.T) {
		for _, ct := range []cache.CacheType{cache.CacheFile, cache.CacheDocker, cache.CacheContainerd, cache.CacheMemory} {
			_, err := NewFileDownloader(ct, nil)
			if err == nil {
				t.Errorf("NewFileDownloader() with cache %v: expected error, got nil", ct)
			}
		}
	})
}

func TestFileBackend_Get(t *testing.T) {
	t.Run("reads file contents via callback", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "payload.txt")
		payload := []byte("hello world")
		if err := os.WriteFile(path, payload, 0o600); err != nil {
			t.Fatalf("setup: WriteFile: %v", err)
		}

		d, err := NewFileDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewFileDownloader() unexpected error: %v", err)
		}

		var got []byte
		err = d.Get(&Source{Url: "file://" + path}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error {
			b, err := io.ReadAll(reader)
			if err != nil {
				return err
			}
			got = b
			return nil
		})
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if !bytes.Equal(got, payload) {
			t.Errorf("callback contents = %q, want %q", got, payload)
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		dir := t.TempDir()
		missing := filepath.Join(dir, "does-not-exist")

		d, err := NewFileDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewFileDownloader() unexpected error: %v", err)
		}

		called := false
		err = d.Get(&Source{Url: "file://" + missing}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error {
			called = true
			return nil
		})
		if err == nil {
			t.Fatal("Get() expected error for missing file, got nil")
		}
		if called {
			t.Error("callback should not be invoked when file cannot be opened")
		}
	})

	t.Run("propagates callback error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "payload.txt")
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatalf("setup: WriteFile: %v", err)
		}

		d, err := NewFileDownloader(cache.CacheNone, nil)
		if err != nil {
			t.Fatalf("NewFileDownloader() unexpected error: %v", err)
		}

		want := errors.New("callback failed")
		err = d.Get(&Source{Url: "file://" + path}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error {
			return want
		})
		if !errors.Is(err, want) {
			t.Errorf("Get() error = %v, want %v", err, want)
		}
	})
}
