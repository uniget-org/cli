//go:build all || docker

package cache

import (
	"fmt"
	"io"
	"testing"

	"github.com/google/safearchive/tar"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
)

func TestNewDockerCache(t *testing.T) {
	_, err := NewDockerCache()
	if err != nil {
		t.Errorf("failed to create DockerCache: %v", err)
	}
}

func TestDockerCacheGet(t *testing.T) {
	cache, err := NewDockerCache()
	if err != nil {
		t.Fatalf("TestDockerCacheGet(): failed to create DockerCache (%v)", err)
	}

	toolRef := containers.NewToolRef("ghcr.io", "uniget-org/tools", "uniget", "main")
	err = cache.Get(toolRef, func(reader io.ReadCloser) error {
		foundUniget := false
		err = archive.ProcessTarContents(reader, func(reader *tar.Reader, header *tar.Header) error {
			if header.Typeflag == tar.TypeReg && header.Name == "bin/uniget" {
				foundUniget = true
				return nil
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("TestDockerCacheGet(): failed to process tar contents (%w)", err)
		}
		if !foundUniget {
			return fmt.Errorf("TestDockerCacheGet(): expected to find 'uniget' in layer tar, but did not")
		}

		return nil
	})
	if err != nil {
		t.Errorf("DockerCache.Get failed: %v", err)
	}
}
