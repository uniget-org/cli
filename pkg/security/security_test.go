package security

import (
	"testing"
)

func TestVerifySigstoreBundleForArtifactWithMissingBundle(t *testing.T) {
	ok, err := VerifySigstoreBundleForArtifact(
		"nonexistent-artifact",
		"nonexistent.sigstore.json",
		"",
		"",
		"",
		"",
	)
	if err == nil {
		t.Fatalf("expected error for missing bundle, got nil")
	}
	if ok {
		t.Errorf("expected verification failure for missing bundle")
	}
}
