package containers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"testing"

	"github.com/google/safearchive/tar"
)

// buildTestLayerTar builds an in-memory docker save style tar containing
// dummy metadata plus a gzipped layer at blobs/sha256/<sha>.
func buildTestLayerTar(sha string, content []byte) ([]byte, error) {
	buffer := &bytes.Buffer{}
	tarWriter := tar.NewWriter(buffer)

	layerBuffer := &bytes.Buffer{}
	gzipWriter := gzip.NewWriter(layerBuffer)
	if _, err := gzipWriter.Write(content); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	entries := []struct {
		name    string
		content []byte
	}{
		{name: "oci-layout", content: []byte("{\"imageLayoutVersion\":\"1.0.0\"}\n")},
		{name: fmt.Sprintf("blobs/sha256/%s", sha), content: layerBuffer.Bytes()},
	}
	for _, entry := range entries {
		err := tarWriter.WriteHeader(&tar.Header{
			Name:     entry.name,
			Size:     int64(len(entry.content)),
			Typeflag: tar.TypeReg,
		})
		if err != nil {
			return nil, err
		}
		if _, err := tarWriter.Write(entry.content); err != nil {
			return nil, err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func TestUnpackLayerFromDockerImage(t *testing.T) {
	content := []byte("layer content")
	sha := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	tarData, err := buildTestLayerTar(sha, content)
	if err != nil {
		t.Fatalf("failed to build test tar: %s", err)
	}

	called := false
	err = UnpackLayerFromDockerImage(io.NopCloser(bytes.NewReader(tarData)), sha, func(reader io.ReadCloser) error {
		called = true
		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		if !bytes.Equal(data, content) {
			t.Errorf("expected layer content %q, got %q", string(content), string(data))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if !called {
		t.Errorf("expected callback to be called")
	}
}

func TestUnpackLayerFromDockerImageSkipsOtherEntries(t *testing.T) {
	content := []byte("layer content")
	sha := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	tarData, err := buildTestLayerTar("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", content)
	if err != nil {
		t.Fatalf("failed to build test tar: %s", err)
	}

	err = UnpackLayerFromDockerImage(io.NopCloser(bytes.NewReader(tarData)), sha, func(reader io.ReadCloser) error {
		t.Errorf("did not expect callback for non-target layer")
		return nil
	})
	if err == nil {
		t.Fatalf("expected error for missing layer, got nil")
	}
}

func TestUnpackLayerFromDockerImageWithCorruptTar(t *testing.T) {
	err := UnpackLayerFromDockerImage(io.NopCloser(bytes.NewReader([]byte("not a tar"))), "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", func(reader io.ReadCloser) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected error for corrupt tar, got nil")
	}
}
