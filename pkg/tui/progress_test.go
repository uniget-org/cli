package tui

import (
	"io"
	"strings"
	"testing"
)

func TestNewProgressReaderIsQuiet(t *testing.T) {
	reader := NewProgressReader(nil, nil)
	if !reader.IsQuiet() {
		t.Errorf("expected reader with nil callbacks to be quiet")
	}

	reader = NewProgressReader(func(int64) {}, func(int64) {})
	if reader.IsQuiet() {
		t.Errorf("expected reader with callbacks not to be quiet")
	}
}

func TestNewQuietProgressReader(t *testing.T) {
	reader := NewQuietProgressReader()
	if !reader.IsQuiet() {
		t.Errorf("expected quiet reader to be quiet")
	}
}

func TestSetTotalInvokesCallback(t *testing.T) {
	seen := int64(-1)
	reader := NewProgressReader(func(n int64) { seen = n }, nil)

	reader.SetTotal(42)
	if seen != 42 {
		t.Errorf("expected callback to see 42, got %d", seen)
	}
}

func TestReadInvokesProgressCallback(t *testing.T) {
	data := "hello world"
	var progress int64
	reader := NewProgressReader(func(int64) {}, func(n int64) { progress += n })
	reader.SetReader(io.NopCloser(strings.NewReader(data)))

	buf := make([]byte, len(data))
	n, err := reader.Read(buf)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if n != len(data) {
		t.Errorf("expected to read %d bytes, got %d", len(data), n)
	}
	if progress != int64(n) {
		t.Errorf("expected progress callback to see %d, got %d", n, progress)
	}
}

func TestQuietReaderRead(t *testing.T) {
	data := "hello world"
	reader := NewQuietProgressReader()
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

type failingReadCloser struct {
	closeErr error
}

func (f *failingReadCloser) Read([]byte) (int, error) { return 0, io.EOF }

func (f *failingReadCloser) Close() error { return f.closeErr }

func TestClosePropagatesError(t *testing.T) {
	closeErr := io.ErrClosedPipe
	reader := NewQuietProgressReader()
	reader.SetReader(&failingReadCloser{closeErr: closeErr})

	if err := reader.Close(); err != closeErr {
		t.Errorf("expected %v, got %v", closeErr, err)
	}
}
