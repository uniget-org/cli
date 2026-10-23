package archive

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/safearchive/tar"
)

type testTarEntry struct {
	name     string
	typeflag byte
	mode     int64
	content  string
	linkname string
}

func buildTar(t *testing.T, entries []testTarEntry) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	writer := tar.NewWriter(buf)
	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Typeflag: entry.typeflag,
			Mode:     entry.mode,
			Linkname: entry.linkname,
			Size:     int64(len(entry.content)),
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatalf("failed to write tar header: %v", err)
		}
		if header.Typeflag == tar.TypeReg {
			if _, err := writer.Write([]byte(entry.content)); err != nil {
				t.Fatalf("failed to write tar content: %v", err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	return buf.Bytes()
}

func TestUntar(t *testing.T) {
	content := buildTar(t, []testTarEntry{
		{name: "dir", typeflag: tar.TypeDir, mode: 0o755},
		{name: "bin/jq", typeflag: tar.TypeReg, mode: 0o755, content: "binary"},
		{name: "link", typeflag: tar.TypeSymlink, mode: 0o777, linkname: "bin/jq"},
	})

	names := []string{}
	err := Untar(io.NopCloser(bytes.NewReader(content)), func(reader *tar.Reader, header *tar.Header) error {
		names = append(names, header.Name)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"dir", "bin/jq", "link"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(names))
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("expected entry %d to be %s, got %s", i, name, names[i])
		}
	}
}

func TestUntarWithCallbackError(t *testing.T) {
	content := buildTar(t, []testTarEntry{
		{name: "bin/jq", typeflag: tar.TypeReg, mode: 0o755, content: "binary"},
	})
	callbackErr := errors.New("callback failed")

	err := Untar(io.NopCloser(bytes.NewReader(content)), func(reader *tar.Reader, header *tar.Header) error {
		return callbackErr
	})
	if err == nil || !strings.Contains(err.Error(), "failed to process item through callback") {
		t.Fatalf("expected callback error, got %v", err)
	}
}

func TestUntarWithInvalidData(t *testing.T) {
	err := Untar(io.NopCloser(bytes.NewReader([]byte("this is not a tar file"))), func(reader *tar.Reader, header *tar.Header) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected error for invalid tar data, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read next item") {
		t.Errorf("expected error to mention reading item, got %v", err)
	}
}

func TestGunzip(t *testing.T) {
	maxByte := &bytes.Buffer{}
	gzipWriter := gzip.NewWriter(maxByte)
	if _, err := gzipWriter.Write([]byte("gunzipped content")); err != nil {
		t.Fatalf("failed to write gzip content: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	err := Gunzip(io.NopCloser(maxByte), func(reader io.Reader) error {
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("failed to read gunzipped data: %v", err)
		}
		if string(data) != "gunzipped content" {
			t.Errorf("expected %q, got %q", "gunzipped content", string(data))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGunzipWithCallbackError(t *testing.T) {
	pipeReader := &bytes.Buffer{}
	gzipWriter := gzip.NewWriter(pipeReader)
	if _, err := gzipWriter.Write([]byte("content")); err != nil {
		t.Fatalf("failed to write gzip content: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	callbackErr := errors.New("callback failed")
	err := Gunzip(io.NopCloser(pipeReader), func(reader io.Reader) error {
		return callbackErr
	})
	if err == nil || !strings.Contains(err.Error(), "failed to execute callback") {
		t.Fatalf("expected callback error, got %v", err)
	}
}

func TestGunzipWithInvalidData(t *testing.T) {
	err := Gunzip(io.NopCloser(bytes.NewReader([]byte("not gzip"))), func(reader io.Reader) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected error for invalid gzip data, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create gzip reader") {
		t.Errorf("expected error to mention gzip reader, got %v", err)
	}
}
