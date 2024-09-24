package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/uniget-org/cli/pkg/tool"
)

func TestInspect(t *testing.T) {
	tool := tool.Tool{
		Name:    "jq",
		Version: "1.7.1",
		Binary:  "jq",
	}
	tools.Tools = append(tools.Tools, tool)

	tt := []struct {
		name        string
		args        []string
		expectErr   error
		expectOut   string
		outContains bool
	}{
		{
			name:      "tool exists",
			args:      []string{"inspect", "foo"},
			expectErr: fmt.Errorf("error getting tool foo"),
			expectOut: "bar",
		},
		{
			name:      "foo",
			args:      []string{"inspect", "jq"},
			expectErr: nil,
			expectOut: "bin/jq" + "\n" +
				"share/man/man1/jq.1" + "\n" +
				"var/lib/uniget/manifests/jq.json" + "\n" +
				"var/lib/uniget/manifests/jq.txt",
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
