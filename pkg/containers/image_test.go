package containers

import (
	"strings"
	"testing"
)

func TestFindNewDigest(t *testing.T) {
	digest, err := FindNewDigest(NewToolRef("ghcr.io", "uniget-org/tools", "jq", "1.7.1").GetRef())
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if !strings.HasPrefix(digest, "sha256:") {
		t.Errorf("expected digest with sha256 prefix, got %s", digest)
	}
}

func TestFindNewDigestWithUnreachableRegistry(t *testing.T) {
	_, err := FindNewDigest(NewToolRef("127.0.0.1:1", "uniget-org/tools", "jq", "1.7.1").GetRef())
	if err == nil {
		t.Fatalf("expected error for unreachable registry, got nil")
	}
}
