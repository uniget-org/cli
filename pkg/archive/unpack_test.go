package archive

import (
	"archive/tar"
	"io"
	"os"
	"testing"
)

var (
	testTarGz = "../../testdata/foo.tar.gz"
	testTar   = "../../testdata/foo.tar.gz"
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

func unpackTestArchive(testArchive string, t *testing.T) string {
	reader := openTestArchive(testArchive, t)

	tempDir := t.TempDir()
	curDir, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get current directory: %v", err)
	}
	err = os.Chdir(tempDir)
	if err != nil {
		t.Errorf("failed to change directory to %s: %v", tempDir, err)
	}

	err = ExtractTarGz(reader, func(path string) string { return path }, func(path string) {})
	if err != nil {
		t.Errorf("failed to extract tar.gz: %v", err)
	}

	err = os.Chdir(curDir)
	if err != nil {
		t.Errorf("failed to change directory to %s: %v", curDir, err)
	}

	return tempDir
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

func TestExtractTarGz(t *testing.T) {
	tempDir := unpackTestArchive(testTarGz, t)

	_, err := os.Stat(tempDir + "/foo")
	if err != nil {
		t.Errorf("expected foo to exist: %v", err)
	}

	data, err := os.ReadFile(tempDir + "/foo")
	if err != nil {
		t.Errorf("failed to read foo: %v", err)
	}
	if string(data) != "bar\n" {
		t.Errorf("expected bar, got %s", string(data))
	}
}

func TestListTarGz(t *testing.T) {
	reader := openTestArchive(testTarGz, t)

	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	if err != nil {
		t.Errorf("failed to change directory to %s: %v", tempDir, err)
	}

	files, err := ListTarGz(reader, func(path string) string { return path })
	if err != nil {
		t.Errorf("failed to extract tar.gz: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	if files[0] != "foo" {
		t.Errorf("expected foo, got %s", files[0])
	}
}

func TestProcessTarContents(t *testing.T) {
	tarData := readTestArchive(testTar, t)

	err := ProcessTarContents(tarData, func(tar *tar.Reader, header *tar.Header) error { return nil })
	if err != nil {
		t.Errorf("failed to process tar contents: %v", err)
	}
}
