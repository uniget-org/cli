package metadata

import (
	"io"
	"strings"
	"testing"
)

func TestNewTarGzUnpacker(t *testing.T) {
	unpacker := NewTarGzUnpacker()
	if unpacker == nil {
		t.Fatal("NewTarGzUnpacker() returned nil interface")
	}

	if _, ok := unpacker.(*TarGzUnpacker); !ok {
		t.Fatalf("NewTarGzUnpacker() returned %T, want *TarGzUnpacker", unpacker)
	}
}

func TestTarGzUnpackerUnpack(t *testing.T) {
	t.Run("returns error for a non-gzip stream", func(t *testing.T) {
		unpacker := TarGzUnpacker{}

		err := unpacker.Unpack(io.NopCloser(strings.NewReader("not-a-real-tarball")))
		if err == nil {
			t.Fatal("Unpack() expected error for non-gzip stream, got nil")
		}
	})
}
