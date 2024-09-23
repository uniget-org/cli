package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/uniget-org/cli/pkg/tool"
)

func TestInspect(t *testing.T) {
	tempDir := t.TempDir()

	tools = tool.Tools{
		tool.Tool{
			Name: "jq",
		},
	}

	tt := []struct {
		name        string
		args        []string
		expectErr   error
		expectOut   string
		outContains bool
	}{
		{
			name:      "foo",
			args:      []string{"inspect", "jq"},
			expectOut: "ocidir://" + tempDir + "testrepo:v2",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, err := cobraTest(t, nil, tc.args...)
			if tc.expectErr != nil {
				if err == nil {
					t.Errorf("did not receive expected error: %v", tc.expectErr)
				} else if !errors.Is(err, tc.expectErr) && err.Error() != tc.expectErr.Error() {
					t.Errorf("unexpected error, received %v, expected %v", err, tc.expectErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("returned unexpected error: %v", err)
			}
			if (!tc.outContains && out != tc.expectOut) || (tc.outContains && !strings.Contains(out, tc.expectOut)) {
				t.Errorf("unexpected output, expected %s, received %s", tc.expectOut, out)
			}
		})
	}
}
