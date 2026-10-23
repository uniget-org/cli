package os

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

var (
	stringForMode = map[int64]string{
		0o777: "rwxrwxrwx",
		0o755: "rwxr-xr-x",
		0o644: "rw-r--r--",
	}
)

func TestFileMode(t *testing.T) {
	for mode, expected := range stringForMode {
		modeString, err := ConvertFileModeToString(mode)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
			continue
		}
		if modeString != expected {
			t.Errorf("expected %s, got %s", expected, modeString)
		}
	}
}

func TestFileModeWithSpecialBits(t *testing.T) {
	tests := []struct {
		name     string
		mode     int64
		expected string
	}{
		{
			name:     "suid bit",
			mode:     0o4755,
			expected: "rwsr-xr-x",
		},
		{
			name:     "sgid bit",
			mode:     0o2755,
			expected: "rwxr-sr-x",
		},
		{
			name:     "sticky bit",
			mode:     0o1755,
			expected: "rwxr-xr-t",
		},
		{
			name:     "no permissions",
			mode:     0o0,
			expected: "---------",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modeString, err := ConvertFileModeToString(tt.mode)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if modeString != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, modeString)
			}
		})
	}
}

func TestFileModeWithUnsupportedMode(t *testing.T) {
	_, err := ConvertFileModeToString(0o17777)
	if err == nil {
		t.Fatalf("expected error for unsupported mode, got nil")
	}
}

func TestSlurpFile(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
		t.Fatalf("failed to write file: %s", err)
	}

	content, err := SlurpFile(filePath)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if !bytes.Equal(content, []byte("hello")) {
		t.Errorf("expected %q, got %q", "hello", string(content))
	}

	_, err = SlurpFile(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Errorf("expected error for nonexistent file, got nil")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "src.txt")
	if err := os.WriteFile(srcPath, []byte("copy me"), 0o600); err != nil {
		t.Fatalf("failed to write source file: %s", err)
	}

	dstPath := filepath.Join(tmpDir, "dst.txt")
	if err := CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	content, err := SlurpFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %s", err)
	}
	if string(content) != "copy me" {
		t.Errorf("expected %q, got %q", "copy me", string(content))
	}

	if err := CopyFile(filepath.Join(tmpDir, "nonexistent"), dstPath); err == nil {
		t.Errorf("expected error for missing source, got nil")
	}
}

func TestCloneFile(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "src.txt")
	if err := os.WriteFile(srcPath, []byte("clone me"), 0o600); err != nil {
		t.Fatalf("failed to write source file: %s", err)
	}

	dstPath := filepath.Join(tmpDir, "dst.txt")
	if err := CloneFile(srcPath, dstPath); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("failed to stat source file: %s", err)
	}
	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("failed to stat destination file: %s", err)
	}

	if dstInfo.Mode() != srcInfo.Mode() {
		t.Errorf("expected mode %s, got %s", srcInfo.Mode(), dstInfo.Mode())
	}
	if !dstInfo.ModTime().Equal(srcInfo.ModTime()) {
		t.Errorf("expected mod time %s, got %s", srcInfo.ModTime(), dstInfo.ModTime())
	}
	if err := CloneFile(filepath.Join(tmpDir, "nonexistent"), dstPath); err == nil {
		t.Errorf("expected error for missing source, got nil")
	}
}

func TestFileExistsAndDirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()

	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to write file: %s", err)
	}

	if !FileExists(filePath) {
		t.Errorf("expected FileExists to be true for %s", filePath)
	}
	if FileExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Errorf("expected FileExists to be false for nonexistent file")
	}
	if !DirectoryExists(tmpDir) {
		t.Errorf("expected DirectoryExists to be true for %s", tmpDir)
	}
	if DirectoryExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Errorf("expected DirectoryExists to be false for nonexistent directory")
	}
}

func TestDirectoryIsWritable(t *testing.T) {
	tmpDir := t.TempDir()

	if !DirectoryIsWritable(tmpDir) {
		t.Errorf("expected %s to be writable", tmpDir)
	}
	if DirectoryIsWritable(filepath.Join(tmpDir, "nonexistent")) {
		t.Errorf("expected nonexistent directory to not be writable")
	}
}
