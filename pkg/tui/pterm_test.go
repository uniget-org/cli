package tui

import (
	"io"
	"strings"
	"testing"
)

func TestNewPtermProgressReader(t *testing.T) {
	reader, err := NewPtermProgressReader("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if reader.IsQuiet() {
		t.Errorf("expected reader not to be quiet")
	}
}

func TestPtermProgressReaderRead(t *testing.T) {
	data := "hello world"
	reader, err := NewPtermProgressReader("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	reader.SetReader(io.NopCloser(strings.NewReader(data)))

	buf := make([]byte, len(data))
	n, err := reader.Read(buf)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if n != len(data) || string(buf) != data {
		t.Errorf("expected to read %s, got %d bytes %s", data, n, string(buf))
	}
}
