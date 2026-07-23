package cache

import (
	"bytes"
	"io"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

func TestNewNoneCache(t *testing.T) {
	c := NewNoneCache()

	if c.Type != CacheNone { // #nosec SA5011 - Test code, c is not nil
		t.Errorf("expected Type %v, got %v", CacheNone, c.Type)
	}
}

func TestNoneCache_GetName(t *testing.T) {
	c := NewNoneCache()

	if got, want := c.GetName(), "none"; got != want {
		t.Errorf("GetName() = %q, want %q", got, want)
	}
}

func TestNoneCache_Put(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		payload []byte
	}{
		{name: "empty payload", key: "empty", payload: []byte{}},
		{name: "non-empty payload", key: "some-key", payload: []byte("hello world")},
		{name: "empty key", key: "", payload: []byte("data")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewNoneCache()
			pr := tui.NewQuietProgressReader()
			reader := io.NopCloser(bytes.NewReader(tt.payload))

			if err := c.Put(tt.key, pr, reader); err != nil {
				t.Errorf("Put() unexpected error: %v", err)
			}
			if c.Has(tt.key) {
				t.Errorf("Has(%q) = true after Put, want false", tt.key)
			}
		})
	}
}

func TestNoneCache_Has(t *testing.T) {
	c := NewNoneCache()

	if c.Has("missing") {
		t.Error("Has(\"missing\") = true, want false")
	}

	pr := tui.NewQuietProgressReader()
	if err := c.Put("stored", pr, io.NopCloser(bytes.NewReader([]byte("data")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}
	if c.Has("stored") {
		t.Error("Has(\"stored\") = true after Put, want false (NoneCache never stores)")
	}
}

func TestNoneCache_Get(t *testing.T) {
	c := NewNoneCache()
	pr := tui.NewQuietProgressReader()

	called := false
	callback := func(reader io.ReadCloser) error {
		called = true
		return nil
	}

	if err := c.Get("missing", pr, callback); err != nil {
		t.Errorf("Get() unexpected error: %v", err)
	}
	if called {
		t.Error("callback was invoked, want it to be skipped for NoneCache")
	}
}

func TestNoneCache_GetAfterPut(t *testing.T) {
	c := NewNoneCache()
	pr := tui.NewQuietProgressReader()

	if err := c.Put("k", pr, io.NopCloser(bytes.NewReader([]byte("value")))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}

	called := false
	err := c.Get("k", pr, func(reader io.ReadCloser) error {
		called = true
		return nil
	})
	if err != nil {
		t.Errorf("Get() unexpected error: %v", err)
	}
	if called {
		t.Error("callback was invoked after Put, want NoneCache to be a no-op")
	}
}

func TestNoneCache_ImplementsCacheInterface(t *testing.T) {
	var _ Cache = NewNoneCache()
}
