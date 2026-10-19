package source

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
	d := NewWebDownloader()
	if d == nil {
		t.Fatal("NewWebDownloader() returned nil")
	}
}

func TestWebBackend_Get(t *testing.T) {
	t.Run("streams response body to callback", func(t *testing.T) {
		var hits int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits++
			_, _ = w.Write([]byte("payload"))
		}))
		defer srv.Close()

		d := NewWebDownloader()

		var got []byte
		err := d.Get(&Source{Url: srv.URL}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error {
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
		if string(got) != "payload" {
			t.Errorf("callback contents = %q, want %q", got, "payload")
		}
		if hits != 1 {
			t.Errorf("http server hits = %d, want 1", hits)
		}
	})

	t.Run("http error", func(t *testing.T) {
		// Reserve then close a listener to obtain a definitely-unused address.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		badURL := srv.URL
		srv.Close()

		d := NewWebDownloader()

		called := false
		err := d.Get(&Source{Url: badURL}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error {
			called = true
			return nil
		})
		if err == nil {
			t.Fatal("Get() expected error for unreachable server, got nil")
		}
		if called {
			t.Error("callback should not be invoked when the request fails")
		}
	})

	t.Run("propagates callback error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("payload"))
		}))
		defer srv.Close()

		d := NewWebDownloader()

		want := errors.New("callback failed")
		err := d.Get(&Source{Url: srv.URL}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error {
			return want
		})
		if !errors.Is(err, want) {
			t.Errorf("Get() error = %v, want %v", err, want)
		}
	})
}
