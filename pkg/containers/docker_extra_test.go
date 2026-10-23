package containers

import (
	"io"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

func skipIfNoDocker(t *testing.T) {
	t.Helper()
	if !DockerIsAvailable() {
		t.Skipf("docker daemon not reachable")
	}
}

func getDockerClientForTest(t *testing.T) (clean func()) {
	t.Helper()
	cli, err := GetDockerClient()
	if err != nil {
		t.Fatalf("failed to get docker client: %s", err)
	}
	return func() {
		_ = cli.Close()
	}
}

func TestDockerClientAndAvailability(t *testing.T) {
	skipIfNoDocker(t)

	closer := getDockerClientForTest(t)
	defer closer()
	closer()
}

func TestListDockerImagesByPrefixNotFound(t *testing.T) {
	skipIfNoDocker(t)

	cli, err := GetDockerClient()
	if err != nil {
		t.Fatalf("failed to get docker client: %s", err)
	}
	defer func() {
		_ = cli.Close()
	}()

	images, err := ListDockerImagesByPrefix(cli, "uniget-test-nonexistent-prefix")
	if err != nil {
		t.Fatalf("failed to list images: %s", err)
	}
	if len(images) != 0 {
		t.Errorf("expected no images for unknown prefix, got %d", len(images))
	}

	if err := RemoveDockerImage(cli, "uniget-test-nonexistent-prefix/base:latest"); err == nil {
		t.Errorf("expected error removing unknown image, got nil")
	}
}

func TestDockerPullAndReadFromLocalRegistry(t *testing.T) {
	skipIfNoDocker(t)

	cli, err := GetDockerClient()
	if err != nil {
		t.Fatalf("failed to get docker client: %s", err)
	}
	defer func() {
		_ = cli.Close()
	}()
	image := toolRef.String()

	if err := PullDockerImage(cli, image); err != nil {
		t.Fatalf("failed to pull image %s from local registry: %s", image, err)
	}
	if !CheckDockerImageExists(cli, image) {
		t.Errorf("expected image %s to exist after pull", image)
	}
	defer func() {
		_ = RemoveDockerImage(cli, image)
	}()

	if err := PullDockerImage(cli, "this is invalid"); err == nil {
		t.Errorf("expected error for invalid image reference, got nil")
	}

	progress := tui.NewProgressReader(nil, nil)
	var payloadSize int64
	err = ReadDockerImage(cli, image, progress, func(reader io.ReadCloser) error {
		size, err := io.Copy(io.Discard, reader)
		if err != nil {
			t.Fatalf("failed to discard image payload: %s", err)
		}
		payloadSize = size
		return nil
	})
	if err != nil {
		t.Fatalf("failed to read docker image: %s", err)
	}
	if payloadSize == 0 {
		t.Errorf("expected image payload to be non-empty")
	}
}

func TestGetFirstLayerFromDockerImageFromLocalRegistry(t *testing.T) {
	skipIfNoDocker(t)

	cli, err := GetDockerClient()
	if err != nil {
		t.Fatalf("failed to get docker client: %s", err)
	}
	defer func() {
		_ = cli.Close()
	}()

	toolRefLocal := NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
	image := toolRefLocal.String()
	if err := PullDockerImage(cli, image); err != nil {
		t.Fatalf("failed to pull image %s: %s", image, err)
	}
	defer func() {
		_ = RemoveDockerImage(cli, image)
	}()

	progress := tui.NewProgressReader(nil, nil)
	err = GetFirstLayerFromDockerImage(cli, toolRefLocal, progress, func(reader io.ReadCloser) error {
		size, err := io.Copy(io.Discard, reader)
		if err != nil {
			t.Fatalf("failed to read layer: %s", err)
		}
		if size == 0 {
			t.Errorf("expected layer to be non-empty")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to get first layer from docker image: %s", err)
	}
}
