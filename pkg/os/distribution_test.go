package os

import (
	"os"
	"testing"
)

func TestGetOsVendor(t *testing.T) {
	tempDir := t.TempDir()

	err := os.MkdirAll(tempDir+"/etc", 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(tempDir+"/etc/os-release", []byte("ID=\"debian\""), 0644)
	if err != nil {
		t.Fatal(err)
	}
	vendor, err := GetOsVendor(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if vendor != "debian" {
		t.Errorf("expected debian, got %s", vendor)
	}
}

func TestGetOsVendorFromMissingFile(t *testing.T) {
	tempDir := t.TempDir()

	err := os.WriteFile(tempDir+"/etc/os-release", []byte("ID=\"debian\""), 0644)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
