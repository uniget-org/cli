package parse

import (
	"os"
	"reflect"
	"testing"

	"github.com/regclient/regclient/types/ref"
)

var parseComposeFileTestCases = []struct {
	name         string
	files        map[string]string
	expected     []string
	expectedRefs ImageRefs
}{
	{
		name: "Single compose",
		files: map[string]string{
			"compose.yaml": "services:\n" +
				"  foo:\n" +
				"    image: alpine:latest",
		},
		expected: []string{
			"alpine:latest",
		},
	},
	{
		name: "Single compose",
		files: map[string]string{
			"compose.yaml": "services:\n" +
				"  foo:\n" +
				"    build:\n" +
				"      context: .\n" +
				"      dockerfile: Dockerfile",
			"Dockerfile": "FROM alpine:latest",
		},
		expected: []string{},
	},
}

func TestExtractImageReferencesFromComposeFile(t *testing.T) {
	for _, tc := range parseComposeFileTestCases {
		t.Run(tc.name, func(t *testing.T) {

			tc.expectedRefs.Refs = make([]ref.Ref, len(tc.expected))
			for i, s := range tc.expected {
				r, err := ref.New(s)
				if err != nil {
					t.Fatalf("Failed to create ref from %q: %v", s, err)
				}
				tc.expectedRefs.Refs[i] = r
			}

			tempDir := t.TempDir()
			curDir, err := os.Getwd()
			if err != nil {
				t.Errorf("failed to get current directory: %v", err)
			}
			err = os.Chdir(tempDir)
			if err != nil {
				t.Errorf("failed to change directory to %s: %v", tempDir, err)
			}

			for name, content := range tc.files {
				err := os.WriteFile(name, []byte(content), 0644) // #nosec G306 -- Test with temporary directory
				if err != nil {
					t.Fatalf("Failed to write file %q: %v", name, err)
				}
			}

			var imageRefs ImageRefs
			project, err := LoadComposeFile("compose.yaml")
			if err != nil {
				t.Errorf("failed to load compose file: %v", err)
			}
			imageRefs, err = ExtractImageReferencesFromComposeFile(project)
			if err != nil {
				t.Errorf("failed to extract image references: %v", err)
			}

			err = os.Chdir(curDir)
			if err != nil {
				t.Errorf("failed to change directory to %s: %v", curDir, err)
			}

			if len(imageRefs.Refs) > 0 && len(tc.expectedRefs.Refs) > 0 {
				if !reflect.DeepEqual(imageRefs.Refs, tc.expectedRefs.Refs) {
					t.Errorf("\nExpected %+v\ngot      %+v", tc.expectedRefs, imageRefs)
				}
			}
		})
	}
}
