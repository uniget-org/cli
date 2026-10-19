package source

import (
	"io"
	"testing"

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
	d := NewOciDownloader()
	if d == nil {
		t.Fatal("NewOciDownloader() returned nil")
	}
}

func TestOciBackend_Get(t *testing.T) {
	// A wrong-but-valid-scheme ref exercises the full registry path without
	// network: regclient fails creating the reference first.
	d := NewOciDownloader()

	// An uppercase letter in the repository segment is rejected by regclient.
	err := d.Get(&Source{Url: "oci://example.com/BAD:tag"}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
	if err == nil {
		t.Fatal("Get() expected error for invalid ref, got nil")
	}
}

func TestOciBackend_GetFromRegistry_InvalidRef(t *testing.T) {
	d := NewOciDownloader()

	// An uppercase letter in the repository segment is rejected by regclient.
	err := d.GetFromRegistry(&Source{Url: "oci://example.com/BAD:tag"}, tui.NewQuietProgressReader(), func(reader io.ReadCloser) error { return nil })
	if err == nil {
		t.Fatal("GetFromRegistry() expected error for invalid ref, got nil")
	}
}
