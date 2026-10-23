package containers

import (
	"context"
	"errors"
	"io"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

func TestContainerdClientCreation(t *testing.T) {
	// This must work without a running containerd daemon: the client
	// connects lazily.
	client, err := GetContainerdClient()
	if err != nil {
		t.Fatalf("failed to get containerd client: %s", err)
	}
	if client == nil {
		t.Fatalf("expected client, got nil")
	}
	t.Cleanup(func() {
		_ = client.Close()
	})
}

func TestContainerdIsAvailable(t *testing.T) {
	// Only assert the call succeeds; availability depends on the
	// environment.
	_ = ContainerdIsAvailable()
}

func TestContainerdPullFailsForUnknownImage(t *testing.T) {
	client, err := GetContainerdClient()
	if err != nil {
		t.Fatalf("failed to get containerd client: %s", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	if err := PullContainerdImage(client, "registry.invalid/uniget-nonexistent/tool:latest"); err == nil {
		t.Errorf("expected pull to fail for unknown image, got nil")
	}

	if CheckContainerdImageExists(client, "registry.invalid/uniget-nonexistent/tool:latest") {
		t.Errorf("expected image to not exist")
	}

	progress := tui.NewProgressReader(nil, nil)
	err = ReadContainerdImage(client, "registry.invalid/uniget-nonexistent/tool:latest", progress, func(_ io.ReadCloser) error {
		return nil
	})
	if err == nil {
		t.Errorf("expected read to fail, got nil")
	}
}

func TestGetFirstLayerFromContainerdImageFailsForBadImage(t *testing.T) {
	client, err := GetContainerdClient()
	if err != nil {
		t.Fatalf("failed to get containerd client: %s", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	// Bad registry reference: fails while getting the layer sha
	badRef := NewToolRef("127.0.0.1:5001", registryRepository, "nonexistent", "latest")
	progress := tui.NewProgressReader(nil, nil)
	err = GetFirstLayerFromContainerdImage(client, badRef, progress, func(_ io.ReadCloser) error {
		return nil
	})
	if err == nil {
		t.Errorf("expected error for bad image, got nil")
	}

	// Good registry sha but no containerd daemon: pull fails
	err = GetFirstLayerFromContainerdImage(client, toolRef, progress, func(_ io.ReadCloser) error {
		return nil
	})
	if err == nil {
		t.Errorf("expected error for missing daemon, got nil")
	}
}

func TestGetFirstLayerFromRegistryCallbackError(t *testing.T) {
	r := toolRef.GetRef()
	rc := GetRegclient()
	defer func() {
		err := rc.Close(context.Background(), r)
		if err != nil {
			t.Errorf("failed to close ref %s: %v", r, err)
		}
	}()

	progress := tui.NewProgressReader(nil, nil)
	callbackErr := errors.New("callback failed")
	err := GetFirstLayerFromRegistry(context.Background(), rc, r, progress, func(_ io.ReadCloser) error {
		return callbackErr
	})
	if err == nil {
		t.Errorf("expected callback error, got nil")
	}

	err = GetFirstLayerFromRegistryRaw(context.Background(), rc, r, progress, func(_ io.ReadCloser) error {
		return callbackErr
	})
	if err == nil {
		t.Errorf("expected callback error, got nil")
	}
}
