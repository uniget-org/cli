package metadata

import (
	"io"
	"strings"
	"testing"
)

func TestNewTarGzUnpacker(t *testing.T) {
	unpacker := NewTarGzUnpacker()
	if unpacker == nil {
		t.Fatal("NewTarGzUnpacker() returned nil")
	}
	if *unpacker == nil {
		t.Fatal("NewTarGzUnpacker() returned nil interface")
	}

	if _, ok := (*unpacker).(TarGzUnpacker); !ok {
		t.Fatalf("NewTarGzUnpacker() returned %T, want TarGzUnpacker", *unpacker)
	}
}

func TestTarGzUnpackerUnpack(t *testing.T) {
	t.Run("returns nil for a readable stream", func(t *testing.T) {
		unpacker := TarGzUnpacker{}

		err := unpacker.Unpack(io.NopCloser(strings.NewReader("not-a-real-tarball")))
		if err != nil {
			t.Fatalf("Unpack() unexpected error: %v", err)
		}
	})

	t.Run("returns nil for nil reader in current placeholder implementation", func(t *testing.T) {
		unpacker := TarGzUnpacker{}

		err := unpacker.Unpack(nil)
		if err != nil {
			t.Fatalf("Unpack() unexpected error: %v", err)
		}
	})
}
