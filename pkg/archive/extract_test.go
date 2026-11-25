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
		{name: "bin/jq-setgid", typeflag: tar.TypeReg, mode: 0o2755, content: "binary"},
		{name: "tmp/sticky", typeflag: tar.TypeReg, mode: 0o1755, content: "sticky"},
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
	if got := setuidFile.Mode().Perm(); got != 0o755 {
		t.Errorf("expected mode 0755, got %s", got)
	}
	if got := setuidFile.Mode() & (os.ModeSetuid | os.ModeSetgid | os.ModeSticky); got != 0 {
		t.Errorf("expected special mode bits to be stripped, got %s", got)
	}

	for _, path := range []string{"bin/jq-setgid", "tmp/sticky"} {
		file, statErr := os.Stat(path)
		if statErr != nil {
			t.Fatalf("expected extracted file %s to exist: %v", path, statErr)
		}
		if got := file.Mode().Perm(); got != 0o755 {
			t.Errorf("expected mode 0755 for %s, got %s", path, got)
		}
		if got := file.Mode() & (os.ModeSetuid | os.ModeSetgid | os.ModeSticky); got != 0 {
			t.Errorf("expected special mode bits to be stripped for %s, got %s", path, got)
		}
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

func TestCallbackExtractTarItemRejectsSymlinkEscape(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)
	outside := t.TempDir()

	tests := []struct {
		name     string
		typeflag byte
		linkname string
	}{
		{name: "absolute symlink target", typeflag: tar.TypeSymlink, linkname: filepath.Join(outside, "target")},
		{name: "relative symlink escape", typeflag: tar.TypeSymlink, linkname: "../../outside"},
		{name: "absolute hardlink target", typeflag: tar.TypeLink, linkname: filepath.Join(outside, "target")},
		{name: "relative hardlink escape", typeflag: tar.TypeLink, linkname: "../../outside"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := buildTar(t, []testTarEntry{
				{name: "link", typeflag: tt.typeflag, mode: 0o777, linkname: tt.linkname},
			})

			err := Untar(io.NopCloser(bytes.NewReader(content)), CallbackExtractTarItem)
			if err == nil {
				t.Fatalf("expected error for escaping link target, got nil")
			}
			if _, statErr := os.Lstat(filepath.Join(workDir, "link")); statErr == nil {
				t.Errorf("expected link to not be created")
			}
		})
	}
}

func TestCallbackExtractTarItemExtractsContainedLinks(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)

	content := buildTar(t, []testTarEntry{
		{name: "bin/tool", typeflag: tar.TypeReg, mode: 0o755, content: "binary"},
		{name: "bin/tool-link", typeflag: tar.TypeSymlink, mode: 0o777, linkname: "tool"},
		{name: "bin/tool-hard", typeflag: tar.TypeLink, mode: 0o755, linkname: "bin/tool"},
	})

	err := Untar(io.NopCloser(bytes.NewReader(content)), CallbackExtractTarItem)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	target, err := os.Readlink(filepath.Join(workDir, "bin", "tool-link"))
	if err != nil {
		t.Fatalf("expected symlink to exist: %v", err)
	}
	if target != "tool" {
		t.Errorf("expected symlink target %q, got %q", "tool", target)
	}

	hardInfo, err := os.Lstat(filepath.Join(workDir, "bin", "tool-hard"))
	if err != nil {
		t.Fatalf("expected hard link to exist: %v", err)
	}
	if hardInfo.Mode()&os.ModeSymlink != 0 {
		t.Errorf("expected hard link, got symlink")
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

func TestExtractFileFromTarRejectsNegativeMode(t *testing.T) {
	workDir := t.TempDir()

	header := &tar.Header{Name: "evil", Typeflag: tar.TypeReg, Mode: -1}
	err := ExtractFileFromTar(workDir, header.Name, tar.NewReader(bytes.NewReader(nil)), header)
	if err == nil {
		t.Fatalf("expected error for negative mode, got nil")
	}
	if _, statErr := os.Stat(filepath.Join(workDir, "evil")); statErr == nil {
		t.Errorf("expected file to not be created for negative mode")
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
