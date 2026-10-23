package archive

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/safearchive/tar"
)

func TestCallbackExtractTarItemOffline(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)

	entries := []testTarEntry{
		{name: "share/man/man1/jq.1", typeflag: tar.TypeReg, mode: 0o644, content: "manual"},
		{name: "bin/jq", typeflag: tar.TypeReg, mode: 0o4755, content: "binary"},
		{name: "dir", typeflag: tar.TypeDir, mode: 0o755},
		{name: "bin/jq-link", typeflag: tar.TypeSymlink, mode: 0o777, linkname: "bin/jq"},
		{name: "bin/jq-other", typeflag: tar.TypeLink, mode: 0o777, linkname: "bin/jq"},
		{name: "char-dev", typeflag: tar.TypeChar, mode: 0o666},
	}

	runExtract := func() {
		err := Untar(io.NopCloser(bytes.NewReader(buildTar(t, entries))), CallbackExtractTarItem)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	// First extraction creates files and symlinks
	runExtract()

	regFile, err := os.Stat(filepath.Join("share", "man", "man1", "jq.1"))
	if err != nil {
		t.Fatalf("expected extracted file to exist: %v", err)
	}
	if got := regFile.Mode().Perm(); got != 0o644 {
		t.Errorf("expected mode 0644, got %s", got)
	}

	setuidFile, err := os.Stat("bin/jq")
	if err != nil {
		t.Fatalf("expected extracted setuid-candidate file to exist: %v", err)
	}
	// The setuid warning branch was exercised; the exact mode can be
	// platform dependent (stripped on nosuid mounts), so only the
	// permission bits are asserted.
	if got := setuidFile.Mode().Perm(); got != 0o755 {
		t.Errorf("expected mode 0755, got %s", got)
	}

	// Second extraction hits the "symlink already exists" branch
	runExtract()

	linkTarget, err := os.Readlink("bin/jq-link")
	if err != nil {
		t.Fatalf("expected symlink to exist: %v", err)
	}
	if linkTarget != "bin/jq" {
		t.Errorf("expected symlink target %q, got %q", "bin/jq", linkTarget)
	}
}

func TestCallbackExtractTarItemRejectsEscape(t *testing.T) {
	// safearchive sanitizes traversal names when reading the header,
	// so extraction must not fail. It must never write outside the
	// current directory though.
	workDir := t.TempDir()
	t.Chdir(workDir)

	content := buildTar(t, []testTarEntry{
		{name: "../outside.txt", typeflag: tar.TypeReg, mode: 0o644, content: "payload"},
	})

	err := Untar(io.NopCloser(bytes.NewReader(content)), CallbackExtractTarItem)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, statErr := os.Stat(workDir + "/../outside.txt"); statErr == nil {
		t.Errorf("expected file to not be created outside workDir")
	}
}

func TestCallbackDisplayTarItemOffline(t *testing.T) {
	entries := []testTarEntry{
		{name: "dir", typeflag: tar.TypeDir, mode: 0o755},
		{name: "bin/jq", typeflag: tar.TypeReg, mode: 0o755, content: "binary"},
		{name: "share/man/man1/jq.1", typeflag: tar.TypeSymlink, mode: 0o777, linkname: "jq.9"},
		{name: "bin/jq-gold", typeflag: tar.TypeLink, mode: 0o755, linkname: "bin/jq"},
		{name: "char-dev", typeflag: tar.TypeChar, mode: 0o666},
	}

	err := Untar(io.NopCloser(bytes.NewReader(buildTar(t, entries))), CallbackDisplayTarItem)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExtractFileFromTarOffline(t *testing.T) {
	workDir := t.TempDir()

	content := buildTar(t, []testTarEntry{
		{name: "subdir/somefile", typeflag: tar.TypeReg, mode: 0o600, content: "payload"},
	})

	tarReader := tar.NewReader(bytes.NewReader(content))
	header, err := tarReader.Next()
	if err != nil {
		t.Fatalf("failed to read tar header: %v", err)
	}

	err = ExtractFileFromTar(workDir, header.Name, tarReader, header)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	extracted, statErr := os.Stat(filepath.Join(workDir, "subdir", "somefile"))
	if statErr != nil {
		t.Fatalf("expected extracted file to exist: %v", statErr)
	}
	if got := extracted.Mode().Perm(); got != 0o600 {
		t.Errorf("expected mode 0600, got %s", got)
	}
}

func TestExtractFileFromTarRejectsEscape(t *testing.T) {
	workDir := t.TempDir()

	entries := buildTar(t, []testTarEntry{
		{name: "innocent", typeflag: tar.TypeReg, mode: 0o600, content: "payload"},
	})

	tarReader := tar.NewReader(bytes.NewReader(entries))
	header, err := tarReader.Next()
	if err != nil {
		t.Fatalf("failed to read tar header: %v", err)
	}

	err = ExtractFileFromTar(workDir, "../outside.txt", tarReader, header)
	if err == nil {
		t.Fatalf("expected error for escaping path, got nil")
	}
	if _, statErr := os.Stat(workDir + "/../outside.txt"); statErr == nil {
		t.Errorf("expected file to not be created outside workDir")
	}
}
