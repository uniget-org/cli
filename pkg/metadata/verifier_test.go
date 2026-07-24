package metadata

import (
	"strings"
	"testing"
)

func TestNewSigstoreMetadataVerifier(t *testing.T) {
	verifier := NewSigstoreMetadataVerifier(
		"issuer-value",
		"issuer-regex",
		"san-value",
		"san-regex",
	)
	if verifier == nil {
		t.Fatal("NewSigstoreMetadataVerifier() returned nil")
	}
	if *verifier == nil {
		t.Fatal("NewSigstoreMetadataVerifier() returned nil interface")
	}

	sigstoreVerifier, ok := (*verifier).(SigstoreMetadataVerifier)
	if !ok {
		t.Fatalf("NewSigstoreMetadataVerifier() returned %T, want SigstoreMetadataVerifier", *verifier)
	}

	if got, want := sigstoreVerifier.issuer, "issuer-value"; got != want {
		t.Errorf("issuer = %q, want %q", got, want)
	}
	if got, want := sigstoreVerifier.issuerRegex, "issuer-regex"; got != want {
		t.Errorf("issuerRegex = %q, want %q", got, want)
	}
	if got, want := sigstoreVerifier.san, "san-value"; got != want {
		t.Errorf("san = %q, want %q", got, want)
	}
	if got, want := sigstoreVerifier.sanRegex, "san-regex"; got != want {
		t.Errorf("sanRegex = %q, want %q", got, want)
	}
}

func TestSigstoreMetadataVerifierVerify(t *testing.T) {
	t.Run("wraps bundle loading errors", func(t *testing.T) {
		verifier := SigstoreMetadataVerifier{}

		err := verifier.Verify(&MetadataSource{
			Files: map[string]string{
				"metadata.json":               "/tmp/metadata.json",
				"metadata.json.sigstore.json": "/path/that/does/not/exist.sigstore.json",
			},
		})
		if err == nil {
			t.Fatal("Verify() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "error verifying sigstore bundle for metadata") {
			t.Fatalf("Verify() error = %q, want wrapped verifier error", err)
		}
		if !strings.Contains(err.Error(), "error loading bundle from path") {
			t.Fatalf("Verify() error = %q, want bundle loading error", err)
		}
	})

	t.Run("missing file entries still returns wrapped error", func(t *testing.T) {
		verifier := SigstoreMetadataVerifier{}

		err := verifier.Verify(&MetadataSource{Files: map[string]string{}})
		if err == nil {
			t.Fatal("Verify() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "error verifying sigstore bundle for metadata") {
			t.Fatalf("Verify() error = %q, want wrapped verifier error", err)
		}
	})
}
