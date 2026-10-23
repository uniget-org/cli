package security

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// skipIfNoNetwork skips online integration tests when the Sigstore TUF
// repository cannot be reached.
func skipIfNoNetwork(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head("https://tuf-repo-cdn.sigstore.dev/")
	if err != nil {
		t.Skipf("skipping online test, network unavailable: %s", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
}

const minimalValidBundleJSON = `{
	"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.1",
	"verificationMaterial": {
		"publicKey": {
			"hint": "test"
		}
	},
	"messageSignature": {
		"messageDigest": {
			"algorithm": "SHA2_256",
			"digest": "AA=="
		}
	}
}`

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

func TestVerifySigstoreBundleForArtifactErrorPaths(t *testing.T) {
	tmpDir := t.TempDir()

	invalidBundlePath := filepath.Join(tmpDir, "invalid.sigstore.json")
	if err := os.WriteFile(invalidBundlePath, []byte("not json"), 0o600); err != nil {
		t.Fatalf("failed to write invalid bundle fixture: %s", err)
	}

	validBundlePath := filepath.Join(tmpDir, "valid.sigstore.json")
	if err := os.WriteFile(validBundlePath, []byte(minimalValidBundleJSON), 0o600); err != nil {
		t.Fatalf("failed to write valid bundle fixture: %s", err)
	}

	tests := []struct {
		name        string
		artifact    string
		bundle      string
		expectedSub string
	}{
		{
			name:        "invalid bundle json",
			artifact:    "nonexistent-artifact",
			bundle:      invalidBundlePath,
			expectedSub: "error loading bundle from path",
		},
		{
			name:        "valid bundle with missing artifact",
			artifact:    filepath.Join(tmpDir, "nonexistent-artifact"),
			bundle:      validBundlePath,
			expectedSub: "error opening artifact file",
		},
		{
			name:        "bundle path is a directory",
			artifact:    "nonexistent-artifact",
			bundle:      tmpDir,
			expectedSub: "error loading bundle from path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := VerifySigstoreBundleForArtifact(
				tt.artifact,
				tt.bundle,
				"",
				"",
				"",
				"",
			)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if ok {
				t.Errorf("expected verification failure")
			}
			if !strings.Contains(err.Error(), tt.expectedSub) {
				t.Errorf("expected error containing %q, got %q", tt.expectedSub, err.Error())
			}
		})
	}
}

func TestVerifySignstoreBundleForContainerImageWithInvalidReference(t *testing.T) {
	tests := []struct {
		name  string
		image string
	}{
		{
			name:  "empty image",
			image: "",
		},
		{
			name:  "image with spaces",
			image: "valid name with spaces",
		},
		{
			name:  "image with invalid tag",
			image: "registry.example.com/foo:tag@sha256:tooshort",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := VerifySignstoreBundleForContainerImage(
				tt.image,
				"",
				"",
				"",
				"",
			)
			if err == nil {
				t.Fatalf("expected error for invalid image reference %q, got nil", tt.image)
			}
			if ok {
				t.Errorf("expected verification failure for invalid image reference %q", tt.image)
			}
		})
	}
}

func TestGetSigstoreTrustedRootOnline(t *testing.T) {
	skipIfNoNetwork(t)

	trustedRoot, err := GetSigstoreTrustedRoot()
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if trustedRoot == nil {
		t.Fatalf("expected trusted root, got nil")
	}
}

func TestVerifySigstoreBundleForArtifactOnline(t *testing.T) {
	skipIfNoNetwork(t)

	tmpDir := t.TempDir()

	bundlePath := filepath.Join(tmpDir, "bundle.sigstore.json")
	if err := os.WriteFile(bundlePath, []byte(minimalValidBundleJSON), 0o600); err != nil {
		t.Fatalf("failed to write bundle fixture: %s", err)
	}

	artifactPath := filepath.Join(tmpDir, "artifact")
	if err := os.WriteFile(artifactPath, []byte("artifact"), 0o600); err != nil {
		t.Fatalf("failed to write artifact fixture: %s", err)
	}

	tests := []struct {
		name        string
		issuer      string
		issuerRegex string
		san         string
		sanRegex    string
		expectedSub string
	}{
		{
			name:        "invalid san regex",
			sanRegex:    "([invalid",
			expectedSub: "error creating short certificate identity",
		},
		{
			name:        "valid identity with fake bundle",
			issuer:      "https://issuer.test",
			san:         "someone@test.example",
			sanRegex:    ".*",
			expectedSub: "error verifying bundle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := VerifySigstoreBundleForArtifact(
				artifactPath,
				bundlePath,
				tt.issuer,
				tt.issuerRegex,
				tt.san,
				tt.sanRegex,
			)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if ok {
				t.Errorf("expected verification failure")
			}
			if !strings.Contains(err.Error(), tt.expectedSub) {
				t.Errorf("expected error containing %q, got %q", tt.expectedSub, err.Error())
			}
		})
	}
}

func TestVerifySignstoreBundleForContainerImageOnline(t *testing.T) {
	skipIfNoNetwork(t)

	// The image exists and has attestations, but not signed by our test
	// identity, so verification must fail without panicking.
	ok, err := VerifySignstoreBundleForContainerImage(
		"ghcr.io/uniget-org/tools/metadata:main",
		"https://invalid-issuer.test",
		"",
		"invalid@test.example",
		"",
	)
	if err == nil {
		t.Fatalf("expected verification failure for mismatched identity, got success")
	}
	if ok {
		t.Errorf("expected verification failure")
	}
}
