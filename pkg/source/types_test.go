package source

import (
	"fmt"
	"testing"
)

func TestNewBackendFromScheme(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    any
		wantErr bool
	}{
		// want is a zero-value backend of the expected type; only the type is compared.
		{name: "file scheme", url: "file:///tmp/foo", want: &FileBackend{}},
		{name: "http scheme", url: "http://example.com/foo", want: &WebBackend{}},
		{name: "https scheme", url: "https://example.com/foo", want: &WebBackend{}},
		{name: "oci scheme", url: "oci://example.com/foo:tag", want: &OciBackend{}},
		{name: "unsupported scheme", url: "unsupported://example.com/foo", wantErr: true},
		{name: "no scheme", url: "example.com/foo:tag", wantErr: true},
		{name: "empty", url: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewBackendFromScheme(&Source{Url: tt.url})
			if tt.wantErr {
				if err == nil {
					t.Fatal("NewBackendFromScheme() expected error, got nil")
				}

				return
			}
			if err != nil {
				t.Fatalf("NewBackendFromScheme() unexpected error: %v", err)
			}
			if d == nil {
				t.Fatal("NewBackendFromScheme() returned nil downloader")
			}
			if gotType, wantType := typeOf(d), typeOf(tt.want); gotType != wantType {
				t.Errorf("NewBackendFromScheme(%q) returned %T, want %s", tt.url, d, wantType)
			}
		})
	}
}

func typeOf(v any) string {
	switch v.(type) {
	case *FileBackend:
		return "*source.FileBackend"
	case *WebBackend:
		return "*source.WebBackend"
	case *OciBackend:
		return "*source.OciBackend"
	default:
		return fmt.Sprintf("%T", v)
	}
}
