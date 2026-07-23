package cache

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

func TestNewFileCache(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")

	c, err := NewFileCache(dir, 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	if c.Type != CacheFile {
		t.Errorf("expected Type %v, got %v", CacheFile, c.Type)
	}
	if c.cacheDirectory != dir {
		t.Errorf("cacheDirectory = %q, want %q", c.cacheDirectory, dir)
	}
	if c.retentionSeconds != 60 {
		t.Errorf("retentionSeconds = %d, want 60", c.retentionSeconds)
	}
	if stat, err := os.Stat(dir); err != nil {
		t.Errorf("cache directory not created: %v", err)
	} else if !stat.IsDir() {
		t.Errorf("cache path is not a directory")
	}
}

func TestNewFileCache_ExistingDirectory(t *testing.T) {
	dir := t.TempDir()

	c, err := NewFileCache(dir, 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error on existing dir: %v", err)
	}
	if c == nil {
		t.Fatal("NewFileCache() returned nil cache")
	}
}

func TestNewFileCache_MkdirFailure(t *testing.T) {
	parent := t.TempDir()
	blocker := filepath.Join(parent, "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("setup: WriteFile: %v", err)
	}

	// Trying to create a directory below a regular file must fail.
	_, err := NewFileCache(filepath.Join(blocker, "sub"), 60)
	if err == nil {
		t.Fatal("NewFileCache() expected error when MkdirAll fails, got nil")
	}
}

func TestFileCache_GetName(t *testing.T) {
	c, err := NewFileCache(t.TempDir(), 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	if got, want := c.GetName(), "file"; got != want {
		t.Errorf("GetName() = %q, want %q", got, want)
	}
}

func TestFileCache_Put(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		payload []byte
	}{
		{name: "empty payload", key: "empty", payload: []byte{}},
		{name: "non-empty payload", key: "key1", payload: []byte("hello world")},
		{name: "binary payload", key: "bin", payload: []byte{0x00, 0x01, 0x02, 0xff}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			c, err := NewFileCache(dir, 60)
			if err != nil {
				t.Fatalf("NewFileCache() unexpected error: %v", err)
			}
			pr := tui.NewQuietProgressReader()
			reader := io.NopCloser(bytes.NewReader(tt.payload))

			if err := c.Put(tt.key, pr, reader); err != nil {
				t.Fatalf("Put() unexpected error: %v", err)
			}

			got, err := os.ReadFile(filepath.Join(dir, tt.key)) // #nosec G304 - Test code
			if err != nil {
				t.Fatalf("ReadFile() unexpected error: %v", err)
			}
			if !bytes.Equal(got, tt.payload) {
				t.Errorf("file contents = %v, want %v", got, tt.payload)
			}
			if !c.Has(tt.key) {
				t.Errorf("Has(%q) = false after Put, want true", tt.key)
			}
		})
	}
}

func TestFileCache_Put_CreateFailure(t *testing.T) {
	dir := t.TempDir()
	c, err := NewFileCache(dir, 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()

	// A key containing a path separator points into a nonexistent subdirectory,
	// which makes os.Create fail.
	err = c.Put("missing/subdir/key", pr, io.NopCloser(bytes.NewReader([]byte("data"))))
	if err == nil {
		t.Fatal("Put() expected error for invalid path, got nil")
	}
}

// errorReader is an io.ReadCloser that always errors on Read.
type fileErrorReader struct{}

func (fileErrorReader) Read(_ []byte) (int, error) { return 0, errors.New("read failure") }
func (fileErrorReader) Close() error               { return nil }

func TestFileCache_Put_ReaderError(t *testing.T) {
	dir := t.TempDir()
	c, err := NewFileCache(dir, 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()

	if err := c.Put("key", pr, fileErrorReader{}); err == nil {
		t.Fatal("Put() expected error from failing reader, got nil")
	}
}

func TestFileCache_Put_Overwrite(t *testing.T) {
	dir := t.TempDir()
	c, err := NewFileCache(dir, 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()

	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader([]byte("first")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}
	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader([]byte("second")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "k")) // #nosec G304 - Test code
	if err != nil {
		t.Fatalf("ReadFile() unexpected error: %v", err)
	}
	if want := []byte("second"); !bytes.Equal(got, want) {
		t.Errorf("file contents = %q, want %q", got, want)
	}
}

func TestFileCache_Has_Missing(t *testing.T) {
	c, err := NewFileCache(t.TempDir(), 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	if c.Has("missing") {
		t.Error("Has(\"missing\") = true, want false")
	}
}

func TestFileCache_Has_Expired(t *testing.T) {
	dir := t.TempDir()
	c, err := NewFileCache(dir, 1)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()

	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader([]byte("data")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}

	// Backdate the file so it is considered expired.
	past := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(filepath.Join(dir, "k"), past, past); err != nil {
		t.Fatalf("Chtimes() unexpected error: %v", err)
	}

	if c.Has("k") {
		t.Error("Has(\"k\") = true for expired entry, want false")
	}
	if _, err := os.Stat(filepath.Join(dir, "k")); !os.IsNotExist(err) {
		t.Errorf("expected expired cache file to be removed, stat err = %v", err)
	}
}

func TestFileCache_Get_Existing(t *testing.T) {
	dir := t.TempDir()
	c, err := NewFileCache(dir, 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()
	payload := []byte("cached value")

	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader(payload))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}

	called := false
	var got []byte
	err = c.Get("k", pr, func(reader io.ReadCloser) error {
		called = true
		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		got = data
		return nil
	})
	if err != nil {
		t.Errorf("Get() unexpected error: %v", err)
	}
	if !called {
		t.Fatal("callback was not invoked")
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("callback received %q, want %q", got, payload)
	}
}

func TestFileCache_Get_Missing(t *testing.T) {
	c, err := NewFileCache(t.TempDir(), 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()

	called := false
	err = c.Get("missing", pr, func(reader io.ReadCloser) error {
		called = true
		return nil
	})
	if err == nil {
		t.Fatal("Get() expected error for missing key, got nil")
	}
	if called {
		t.Error("callback was invoked for missing key, want it to be skipped")
	}
}

func TestFileCache_Get_CallbackError(t *testing.T) {
	dir := t.TempDir()
	c, err := NewFileCache(dir, 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()
	wantErr := errors.New("callback error")

	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader([]byte("data")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}

	err = c.Get("k", pr, func(reader io.ReadCloser) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("Get() error = %v, want %v", err, wantErr)
	}
}

func TestFileCache_ImplementsCacheInterface(t *testing.T) {
	c, err := NewFileCache(t.TempDir(), 60)
	if err != nil {
		t.Fatalf("NewFileCache() unexpected error: %v", err)
	}
	var _ Cache = c
}
