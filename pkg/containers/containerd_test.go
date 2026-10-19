//go:build all || containerd

package containers

import (
	"io"
	"testing"

	"github.com/google/safearchive/tar"

	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func TestGetContainerdClient(t *testing.T) {
	client, err := GetContainerdClient()
	if err != nil {
		t.Fatalf("Failed to get containerd client: %s", err)
	}
	//nolint:errcheck
	defer client.Close()
}

func TestContainerdAvailability(t *testing.T) {
	client, err := GetContainerdClient()
	if err != nil {
		t.Fatalf("Failed to get containerd client: %s", err)
	}
	//nolint:errcheck
	defer client.Close()

	if !ContainerdIsAvailable() {
		t.Fatalf("containerd is not available")
	}
}

func TestReadFirstLayerFromContainerdImage(t *testing.T) {
	client, err := GetContainerdClient()
	if err != nil {
		t.Fatalf("Failed to get containerd client: %s", err)
	}
	//nolint:errcheck
	defer client.Close()

	toolRef := NewToolRef("ghcr.io", "uniget-org/tools", "continue", "latest")
	image := toolRef.String()

	if !CheckContainerdImageExists(client, image) {
		err := PullContainerdImage(client, image)
		if err != nil {
			t.Fatalf("failed to pull containerd image: %s", err)
		}
	}

	err = GetFirstLayerFromContainerdImage(client, toolRef, tui.NewProgressReader(nil, nil), func(reader io.ReadCloser) error {
		err = archive.ProcessTarContents(io.NopCloser(reader), func(tar *tar.Reader, header *tar.Header) error { return nil })
		if err != nil {
			t.Fatalf("failed to process tar: %s", err)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("failed to read containerd image: %s", err)
	}
}
