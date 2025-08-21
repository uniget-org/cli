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

func loadTool(t *testing.T) io.ReadCloser {
	ctx := context.Background()
	r := toolRef.GetRef()
	rc := containers.GetRegclient()
	defer func() {
		err := rc.Close(ctx, r)
		if err != nil {
			t.Errorf("failed to close ref %s: %v", r, err)
		}
	}()
	registryLayer, err := containers.GetFirstLayerFromRegistry(ctx, rc, r)
	if err != nil {
		t.Errorf("failed to get first layer from registry: %v", err)
	}

	return registryLayer
}

func TestProcessTarContents(t *testing.T) {
	registryLayer := loadTool(t)
	err := ProcessTarContents(registryLayer, func(tar *tar.Reader, header *tar.Header) error { return nil })
	if err != nil {
		t.Errorf("failed to process tar contents: %v", err)
	}
}

func TestProcessTarContentsCallback(t *testing.T) {
	registryLayer := loadTool(t)
	files := []string{
		"bin/jq",
		"share/man/man1/jq.1",
		"var/lib/uniget/manifests/jq.json",
		"var/lib/uniget/manifests/jq.txt",
	}
	err := ProcessTarContents(registryLayer, func(reader *tar.Reader, header *tar.Header) error {
		if header.Typeflag != tar.TypeReg {
			return nil
		}

		if !slices.Contains(files, header.Name) {
			t.Errorf("expected %s to be in %v", header.Name, files)
		}

		return nil
	})
	if err != nil {
		t.Errorf("failed to process tar contents: %v", err)
	}
}

func TestProcessTarContentsDisplay(t *testing.T) {
	registryLayer := loadTool(t)
	err := ProcessTarContents(registryLayer, CallbackDisplayTarItem)
	if err != nil {
		t.Errorf("failed to process tar contents: %v", err)
	}
}

func TestProcessTarContentsExtract(t *testing.T) {
	registryLayer := loadTool(t)

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
	err = ProcessTarContents(registryLayer, CallbackExtractTarItem)
	if err != nil {
		t.Errorf("failed to process tar contents: %v", err)
	}
	for _, file := range files {
		_, err := os.Stat(file)
		if err != nil {
			t.Errorf("expected %s to exist: %v", file, err)
		}
	}

	err = os.Chdir(curDir)
	if err != nil {
		t.Errorf("failed to change directory to %s: %v", curDir, err)
	}
}
