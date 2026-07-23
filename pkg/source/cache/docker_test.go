package cache

import (
	"bytes"
	"io"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/tui"
)

// testDockerImage is a tiny, widely-mirrored image used by tests that need a
// real Docker daemon. Kept small so tests remain fast.
const testDockerImage = "docker.io/library/hello-world:latest"

func TestNewDockerCache(t *testing.T) {
	c, err := NewDockerCache()
	if err != nil {
		// client.New(client.FromEnv) parses env vars only; a failure here is
		// caused by malformed DOCKER_* variables in the test environment.
		t.Skipf("NewDockerCache() failed (likely env issue): %v", err)
	}
	if c.Type != CacheDocker {
		t.Errorf("expected Type %v, got %v", CacheDocker, c.Type)
	}
	if c.cli == nil {
		t.Error("expected cli to be initialized")
	}
}

func TestDockerCache_GetName(t *testing.T) {
	c, err := NewDockerCache()
	if err != nil {
		t.Skipf("NewDockerCache() failed: %v", err)
	}
	if got, want := c.GetName(), "docker"; got != want {
		t.Errorf("GetName() = %q, want %q", got, want)
	}
}

func TestDockerCache_ImplementsCacheInterface(t *testing.T) {
	c, err := NewDockerCache()
	if err != nil {
		t.Skipf("NewDockerCache() failed: %v", err)
	}
	var _ Cache = c
}

// TestCacheName_Docker verifies the entry in the CacheName lookup table used
// by CacheStruct.GetName(). This does not require a Docker daemon.
func TestCacheName_Docker(t *testing.T) {
	if got, want := CacheName[CacheDocker], "docker"; got != want {
		t.Errorf("CacheName[CacheDocker] = %q, want %q", got, want)
	}
}

func TestDockerCache_Put(t *testing.T) {
	c, err := NewDockerCache()
	if err != nil {
		t.Fatalf("NewDockerCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()

	// DockerCache.Put ignores the reader; the key is the image reference and
	// Put pulls that image into the local daemon.
	err = c.Put(testDockerImage, pr, io.NopCloser(bytes.NewReader(nil)))
	if err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}
	if !c.Has(testDockerImage) {
		t.Errorf("Has(%q) = false after Put, want true", testDockerImage)
	}
}

func TestDockerCache_Has(t *testing.T) {
	c, err := NewDockerCache()
	if err != nil {
		t.Fatalf("NewDockerCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()

	// Ensure the image is present before asserting on Has.
	if err := c.Put(testDockerImage, pr, io.NopCloser(bytes.NewReader(nil))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}
	if !c.Has(testDockerImage) {
		t.Errorf("Has(%q) = false for pulled image, want true", testDockerImage)
	}
}

func TestDockerCache_Has_Missing(t *testing.T) {
	c, err := NewDockerCache()
	if err != nil {
		t.Fatalf("NewDockerCache() unexpected error: %v", err)
	}

	const missing = "uniget-cache-test/does-not-exist:definitely-not-a-tag"
	if c.Has(missing) {
		t.Errorf("Has(%q) = true for nonexistent image, want false", missing)
	}
}

func TestDockerCache_Get(t *testing.T) {
	c, err := NewDockerCache()
	if err != nil {
		t.Fatalf("NewDockerCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()

	if err := c.Put(testDockerImage, pr, io.NopCloser(bytes.NewReader(nil))); err != nil {
		t.Fatalf("Put() unexpected error: %v", err)
	}

	called := false
	var readBytes int64
	err = c.Get(testDockerImage, pr, func(reader io.ReadCloser) error {
		called = true
		n, err := io.Copy(io.Discard, reader)
		if err != nil {
			return err
		}
		readBytes = n
		return nil
	})
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if !called {
		t.Fatal("callback was not invoked")
	}
	if readBytes == 0 {
		t.Error("callback received 0 bytes from ImageSave stream, want > 0")
	}
}

func TestDockerCache_Get_Missing(t *testing.T) {
	c, err := NewDockerCache()
	if err != nil {
		t.Fatalf("NewDockerCache() unexpected error: %v", err)
	}
	pr := tui.NewQuietProgressReader()

	const missing = "uniget-cache-test/does-not-exist:definitely-not-a-tag"
	called := false
	err = c.Get(missing, pr, func(reader io.ReadCloser) error {
		called = true
		return nil
	})
	if err == nil {
		t.Fatal("Get() expected error for nonexistent image, got nil")
	}
	if called {
		t.Error("callback was invoked for nonexistent image, want it to be skipped")
	}
}
