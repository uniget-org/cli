package parse

import (
	"reflect"
	"strings"
	"testing"

	"github.com/regclient/regclient/types/ref"
)

var parseTestCases = []struct {
	name         string
	input        string
	expected     []string
	expectedRefs []ref.Ref
}{
	{
		name:  "Single FROM statement",
		input: "FROM alpine:latest",
		expected: []string{
			"alpine:latest",
		},
	},
	{
		name:  "Multiple FROM statements",
		input: "FROM alpine:latest\nFROM ubuntu:20.04",
		expected: []string{
			"alpine:latest",
			"ubuntu:20.04",
		},
	},
	{
		name:  "Multiple FROM statements",
		input: "FROM alpine:latest\nFROM ubuntu:20.04\nCOPY --from=ubuntu:24.04 . .",
		expected: []string{
			"alpine:latest",
			"ubuntu:20.04",
			"ubuntu:24.04",
		},
	},
}

func TestExtractImageReferences(t *testing.T) {
	for _, tc := range parseTestCases {
		t.Run(tc.name, func(t *testing.T) {

			tc.expectedRefs = make([]ref.Ref, len(tc.expected))
			for i, s := range tc.expected {
				r, err := ref.New(s)
				if err != nil {
					t.Fatalf("Failed to create ref from %q: %v", s, err)
				}
				tc.expectedRefs[i] = r
			}

			reader := strings.NewReader(tc.input)
			imageRefs, err := ExtractImageReferences(reader)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if !reflect.DeepEqual(imageRefs.Refs, tc.expectedRefs) {
				t.Errorf("\nExpected %+v\ngot      %+v", tc.expectedRefs, imageRefs)
			}
		})
	}
}
