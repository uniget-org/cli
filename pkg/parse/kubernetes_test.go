package parse

/*
apiVersion: v1
kind: Pod
metadata:
  name: image-volume
spec:
  containers:
  - name: shell
    command: ["sleep", "infinity"]
    image: debian
    volumeMounts:
    - name: volume
      mountPath: /volume
  volumes:
  - name: volume
    image:
      reference: quay.io/crio/artifact:v2
      pullPolicy: IfNotPresent
*/

import (
	"reflect"
	"testing"

	"github.com/regclient/regclient/types/ref"
)

var parseKubernetesFileTestCases = []struct {
	name         string
	input        string
	expected     []string
	expectedRefs ImageRefs
}{
	{
		name: "Single pod",
		input: `
apiVersion: v1
kind: Pod
metadata:
  name: image-volume
spec:
  containers:
  - name: shell
    command: ["sleep", "infinity"]
    image: alpine:latest
    volumeMounts:
    - name: volume
      mountPath: /volume
  volumes:
  - name: volume
    image:
      reference: alpine:latest
      pullPolicy: IfNotPresent
        `,
		expected: []string{
			"alpine:latest",
		},
	},
	{
		name: "Multiple pods",
		input: `
apiVersion: v1
kind: Pod
metadata:
  name: image-volume
spec:
  containers:
  - name: shell
    command: ["sleep", "infinity"]
    image: alpine:latest
---
apiVersion: v1
kind: Pod
metadata:
  name: image-volume2
spec:
  containers:
  - name: shell
    command: ["sleep", "infinity"]
    image: ubuntu:24.04
        `,
		expected: []string{
			"alpine:latest",
		},
	},
}

func TestExtractImageReferencesFromKubernetesFile(t *testing.T) {
	for _, tc := range parseKubernetesFileTestCases {
		t.Run(tc.name, func(t *testing.T) {

			tc.expectedRefs.Refs = make([]ref.Ref, len(tc.expected))
			for i, s := range tc.expected {
				r, err := ref.New(s)
				if err != nil {
					t.Fatalf("Failed to create ref from %q: %v", s, err)
				}
				tc.expectedRefs.Refs[i] = r
			}

			var imageRefs ImageRefs
			manifest, err := LoadKubernetesManifest([]byte(tc.input))
			if err != nil {
				t.Errorf("failed to load compose file: %v", err)
			}
			imageRefs, err = ExtractImageReferencesFromKubernetesManifest(manifest)
			if err != nil {
				t.Errorf("failed to extract image references: %v", err)
			}

			if len(imageRefs.Refs) > 0 && len(tc.expectedRefs.Refs) > 0 {
				if !reflect.DeepEqual(imageRefs.Refs, tc.expectedRefs.Refs) {
					t.Errorf("\nExpected %+v\ngot      %+v", tc.expectedRefs, imageRefs)
				}
			}
		})
	}
}
