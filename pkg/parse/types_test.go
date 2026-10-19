package parse

import (
	"bytes"
	"testing"

	"github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

func TestImageRefsAdd(t *testing.T) {
	imageRefs := ImageRefs{}
	reference, err := ref.New("alpine:latest")
	if err != nil {
		t.Fatalf("failed to create ref: %v", err)
	}

	imageRefs.Add(reference)
	if len(imageRefs.Refs) != 1 || imageRefs.Refs[0].Reference != reference.Reference {
		t.Errorf("expected ref to be added, got %+v", imageRefs)
	}
}

func TestImageRefsBumpPassesThroughOtherRefs(t *testing.T) {
	imageRef, err := ref.New("quay.io/example/custom:1.2.3")
	if err != nil {
		t.Fatalf("failed to create ref: %v", err)
	}

	imageRefs := ImageRefs{}
	imageRefs.Add(imageRef)

	var callbackCalls int
	callback := func(toolName string, oldVersion string, newVersion string) {
		callbackCalls++
	}

	err = imageRefs.Bump(&tool.Tools{}, callback)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if len(imageRefs.BumpedRefs) != 1 {
		t.Fatalf("expected 1 bumped ref, got %d", len(imageRefs.BumpedRefs))
	}
	if imageRefs.BumpedRefs[0].Reference != imageRef.Reference {
		t.Errorf("expected unchanged ref, got %+v", imageRefs.BumpedRefs[0])
	}
	if callbackCalls != 0 {
		t.Errorf("expected no callback calls, got %d", callbackCalls)
	}
}

func TestImageRefsBumpWithUnknownTool(t *testing.T) {
	imageRef, err := ref.New("ghcr.io/uniget-org/tools/doesnotexist:1.0.0")
	if err != nil {
		t.Fatalf("failed to create ref: %v", err)
	}

	imageRefs := ImageRefs{}
	imageRefs.Add(imageRef)

	err = imageRefs.Bump(&tool.Tools{}, nil)
	if err == nil {
		t.Fatalf("expected error for unknown tool, got nil")
	}
}

func TestImageRefsReplace(t *testing.T) {
	first, err := ref.New("docker:27-dind")
	if err != nil {
		t.Fatalf("failed to create ref: %v", err)
	}
	second, err := ref.New("registry.example.com/custom:1.0.0")
	if err != nil {
		t.Fatalf("failed to create ref: %v", err)
	}

	imageRefs := ImageRefs{
		Refs:       []ref.Ref{first, second},
		BumpedRefs: []ref.Ref{second, second},
	}

	// Replace does sequential ReplaceAll per ref, so earlier replacements
	// also affect later replacements of the same text.
	file := []byte(first.Reference + "\n" + second.Reference + "\n")
	replaced, err := imageRefs.Replace(file)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	expected := []byte(second.Reference + "\n" + second.Reference + "\n")
	if !bytes.Equal(replaced, expected) {
		t.Errorf("expected %s, got %s", string(expected), string(replaced))
	}
}
