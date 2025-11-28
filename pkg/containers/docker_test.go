//go:build all || docker

package containers

import (
	"io"
	"testing"

	"github.com/google/safearchive/tar"

	"gitlab.com/uniget-org/cli/pkg/archive"
)

func TestDockerClient(t *testing.T) {
	cli, err := GetDockerClient()
	if err != nil {
		t.Fatalf("Failed to get docker client: %s", err)
	}
	//nolint:errcheck
	defer cli.Close()
}

func TestDockerAvailability(t *testing.T) {
	cli, err := GetDockerClient()
	if err != nil {
		t.Fatalf("Failed to get docker client: %s", err)
	}
	//nolint:errcheck
	defer cli.Close()

	if !DockerIsAvailable() {
		t.Fatalf("Docker is not available")
	}
}

func TestReadFirstLayerFromDockerImage(t *testing.T) {
	cli, err := GetDockerClient()
	if err != nil {
		t.Fatalf("Failed to get docker client: %s", err)
	}
	//nolint:errcheck
	defer cli.Close()

	toolRef := NewToolRef("ghcr.io", "uniget-org/tools", "continue", "main")
	image := toolRef.String()

	if !CheckDockerImageExists(cli, image) {
		err := PullDockerImage(cli, image)
		if err != nil {
			t.Fatalf("failed to pull docker image: %s", err)
		}
	}

	shaString, err := GetFirstLayerShaFromRegistry(toolRef)
	if err != nil {
		t.Fatalf("failed to get first layer sha: %s", err)
	}
	sha := shaString[7:]

	err = ReadDockerImage(cli, image, func(reader io.ReadCloser) error {
		err := UnpackLayerFromDockerImage(reader, sha, func(reader io.ReadCloser) error {
			err = archive.ProcessTarContents(io.NopCloser(reader), func(tar *tar.Reader, header *tar.Header) error { return nil })
			if err != nil {
				t.Fatalf("failed to process tar: %s", err)
			}

			return nil
		})
		if err != nil {
			t.Fatalf("failed to unpack layer: %s", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to read docker image: %s", err)
	}
}
