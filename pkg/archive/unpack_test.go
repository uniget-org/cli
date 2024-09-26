package archive

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"slices"
	"testing"

	"github.com/uniget-org/cli/pkg/containers"
)

var (
	testTarGz = "../../testdata/foo.tar.gz"
	testTar   = "../../testdata/foo.tar.gz"

	registryAddress    = "ghcr.io"
	registryRepository = "uniget-org/tools"
	registryImage      = "jq"
	registryTag        = "latest"
	toolRef            = containers.NewToolRef(registryAddress, registryRepository, registryImage, registryTag)
)

func openTestArchive(testArchive string, t *testing.T) io.Reader {
	reader, err := os.Open(testArchive)
	if err != nil {
		t.Errorf("failed to open %s: %v", testTarGz, err)
	}

	return reader
}

func readTestArchive(testArchive string, t *testing.T) []byte {
	reader := openTestArchive(testArchive, t)

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("failed to read %s: %v", testArchive, err)
	}

	return data
}

func loadTool(t *testing.T) []byte {
	ctx := context.Background()
	r := toolRef.GetRef()
	rc := containers.GetRegclient()
	defer rc.Close(ctx, r)
	registryLayer, err := containers.GetFirstLayerFromRegistry(ctx, rc, r)
	if err != nil {
		t.Errorf("failed to get first layer from registry: %v", err)
	}

	return registryLayer
}

func TestPathIsInsideTarget(t *testing.T) {
	var err error

	err = pathIsInsideTarget("/tmp", "/tmp")
	if err != nil {
		t.Errorf("expected /tmp to be inside /tmp")
	}

	err = pathIsInsideTarget("/tmp", "/tmp/foo")
	if err != nil {
		t.Errorf("expected /tmp/foo to be inside /tmp")
	}

	err = pathIsInsideTarget("/tmp/foo", "/tmp")
	if err != nil {
		t.Errorf("expected /tmp/foo not to be inside /tmp")
	}
}

func TestGunzip(t *testing.T) {
	tarGzBytes := readTestArchive(testTarGz, t)
	tarBytes := readTestArchive(testTar, t)

	_, err := Gunzip(tarGzBytes)
	if err != nil {
		t.Errorf("gunzip failed: %v", err)
	}

	if len(tarGzBytes) != len(tarBytes) {
		t.Errorf("expected gunzip to remove data")
	}
	if string(tarGzBytes) != string(tarBytes) {
		t.Errorf("gunzip'ed data does not match expected data")
	}
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
		"var/lib/uniget/manifests/jq.json",
		"var/lib/uniget/manifests/jq.txt",
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
