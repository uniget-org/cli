package main

import (
	"fmt"
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

	tt := []cobraTest{
		{
			name:        "tool exists",
			args:        []string{"inspect", "foo"},
			expectErr:   fmt.Errorf("error getting tool foo"),
			expectOut:   "",
			outContains: false,
		},
		{
			name:      "contents",
			args:      []string{"inspect", "jq", "--raw"},
			expectErr: nil,
			expectOut: "bin/jq" + "\n" +
				"share/man/man1/jq.1" + "\n" +
				"var/lib/uniget/manifests/jq.json" + "\n" +
				"var/lib/uniget/manifests/jq.txt",
			outContains: true,
		},
	}

	runCobraTests(t, tt)
}
