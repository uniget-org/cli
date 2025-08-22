package archive

import (
	"context"
	"io"
	"os"
	"slices"
	"testing"

	"github.com/google/safearchive/tar"

	"github.com/uniget-org/cli/pkg/containers"
)

var (
	registryAddress    = "ghcr.io"
	registryRepository = "uniget-org/tools"
	registryImage      = "jq"
	registryTag        = "latest"
	toolRef            = containers.NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
)

func loadTool(t *testing.T, callback func(reader io.ReadCloser) error) error {
	ctx := context.Background()
	r := toolRef.GetRef()
	rc := containers.GetRegclient()
	defer func() {
		err := rc.Close(ctx, r)
		if err != nil {
			t.Errorf("failed to close ref %s: %v", r, err)
		}
	}()
	err := containers.GetFirstLayerFromRegistry(ctx, rc, r, func(reader io.ReadCloser) error {
		return callback(reader)
	})
	if err != nil {
		t.Errorf("failed to get first layer from registry: %v", err)
	}

	return nil
}

func TestProcessTarContents(t *testing.T) {
	err := loadTool(t, func(reader io.ReadCloser) error {
		return ProcessTarContents(reader, func(tar *tar.Reader, header *tar.Header) error { return nil })
	})
	if err != nil {
		t.Errorf("failed to process tar contents: %v", err)
	}
}

func TestProcessTarContentsCallback(t *testing.T) {
	files := []string{
		"bin/jq",
		"share/man/man1/jq.1",
		"var/lib/uniget/manifests/jq.json",
		"var/lib/uniget/manifests/jq.txt",
	}
	err := loadTool(t, func(reader io.ReadCloser) error {
		return ProcessTarContents(reader, func(reader *tar.Reader, header *tar.Header) error {
			if header.Typeflag != tar.TypeReg {
				return nil
			}

			if !slices.Contains(files, header.Name) {
				t.Errorf("expected %s to be in %v", header.Name, files)
			}

			return nil
		})
	})
	if err != nil {
		t.Errorf("failed to load tool: %v", err)
	}
}

func TestProcessTarContentsDisplay(t *testing.T) {
	err := loadTool(t, func(reader io.ReadCloser) error {
		return ProcessTarContents(reader, CallbackDisplayTarItem)
	})
	if err != nil {
		t.Errorf("failed to load tool: %v", err)
	}
}

func TestProcessTarContentsExtract(t *testing.T) {
	tempDir := t.TempDir()
	curDir, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current directory: %v", err)
	}
	err = os.Chdir(tempDir)
	if err != nil {
		t.Errorf("failed to change directory to %s: %v", tempDir, err)
	}

	files := []string{
		"bin/jq",
		"share/man/man1/jq.1",
	}

	err = loadTool(t, func(reader io.ReadCloser) error {
		err = ProcessTarContents(reader, CallbackExtractTarItem)
		if err != nil {
			t.Errorf("failed to process tar contents: %v", err)
		}
		for _, file := range files {
			_, err := os.Stat(file)
			if err != nil {
				t.Errorf("expected %s to exist: %v", file, err)
			}
		}

		return nil
	})

	err = os.Chdir(curDir)
	if err != nil {
		t.Errorf("failed to change directory to %s: %v", curDir, err)
	}
}
