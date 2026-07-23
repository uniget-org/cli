package cache

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

func TestNewMemoryCache(t *testing.T) {
	c := NewMemoryCache()

	if c.Type != CacheMemory { // #nosec SA5011 - Test code, c is not nil
		t.Errorf("expected Type %v, got %v", CacheMemory, c.Type)
	}
	if c.data == nil {
		t.Error("expected data map to be initialized")
	}
	if len(c.data) != 0 {
		t.Errorf("expected empty data map, got %d entries", len(c.data))
	}
}

func TestMemoryCache_GetName(t *testing.T) {
	c := NewMemoryCache()

	if got, want := c.GetName(), "memory"; got != want {
		t.Errorf("GetName() = %q, want %q", got, want)
	}
}

func TestMemoryCache_Put(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		payload []byte
	}{
		{name: "empty payload", key: "empty", payload: []byte{}},
		{name: "non-empty payload", key: "some-key", payload: []byte("hello world")},
		{name: "empty key", key: "", payload: []byte("data")},
		{name: "binary payload", key: "bin", payload: []byte{0x00, 0x01, 0x02, 0xff}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMemoryCache()
			pr := tui.NewQuietProgressReader()
			reader := io.NopCloser(bytes.NewReader(tt.payload))

			if err := c.Put(tt.key, pr, reader); err != nil {
				t.Fatalf("Put() unexpected error: %v", err)
			}
			if !c.Has(tt.key) {
				t.Errorf("Has(%q) = false after Put, want true", tt.key)
			}
			if got, want := c.data[tt.key], tt.payload; !bytes.Equal(got, want) {
				t.Errorf("stored payload = %v, want %v", got, want)
			}
		})
	}
}

// errorReader is an io.ReadCloser that always returns an error on Read.
type errorReader struct{}

func (errorReader) Read(_ []byte) (int, error) { return 0, errors.New("read failure") }
func (errorReader) Close() error               { return nil }

func TestMemoryCache_Put_ReaderError(t *testing.T) {
	c := NewMemoryCache()
	pr := tui.NewQuietProgressReader()

	err := c.Put("key", pr, errorReader{})
	if err == nil {
		t.Fatal("Put() expected error from failing reader, got nil")
	}
	if c.Has("key") {
		t.Error("Has(\"key\") = true after failed Put, want false")
	}
}

func TestMemoryCache_Put_Overwrite(t *testing.T) {
	c := NewMemoryCache()
	pr := tui.NewQuietProgressReader()

	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader([]byte("first")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}
	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader([]byte("second")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}

	if got, want := c.data["k"], []byte("second"); !bytes.Equal(got, want) {
		t.Errorf("stored payload = %q, want %q", got, want)
	}
}

func TestMemoryCache_Has(t *testing.T) {
	c := NewMemoryCache()

	if c.Has("missing") {
		t.Error("Has(\"missing\") = true, want false")
	}

	pr := tui.NewQuietProgressReader()
	if err := c.Put("stored", pr, io.NopCloser(bytes.NewReader([]byte("data")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}
	if !c.Has("stored") {
		t.Error("Has(\"stored\") = false after Put, want true")
	}
}

func TestMemoryCache_Get_Missing(t *testing.T) {
	c := NewMemoryCache()
	pr := tui.NewQuietProgressReader()

	called := false
	err := c.Get("missing", pr, func(reader io.ReadCloser) error {
		called = true
		return nil
	})
	if err != nil {
		t.Errorf("Get() unexpected error: %v", err)
	}
	if called {
		t.Error("callback was invoked for missing key, want it to be skipped")
	}
}

func TestMemoryCache_Get_Existing(t *testing.T) {
	c := NewMemoryCache()
	pr := tui.NewQuietProgressReader()
	payload := []byte("cached value")

	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader(payload))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}

	called := false
	var got []byte
	err := c.Get("k", pr, func(reader io.ReadCloser) error {
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
		t.Fatal("callback was not invoked for existing key")
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("callback received %q, want %q", got, payload)
	}
}

func TestMemoryCache_Get_CallbackError(t *testing.T) {
	c := NewMemoryCache()
	pr := tui.NewQuietProgressReader()
	wantErr := errors.New("callback error")

	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader([]byte("data")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}

	err := c.Get("k", pr, func(reader io.ReadCloser) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("Get() error = %v, want %v", err, wantErr)
	}
}

func TestMemoryCache_ImplementsCacheInterface(t *testing.T) {
	var _ Cache = NewMemoryCache()
}
