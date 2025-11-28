package main

import (
	"fmt"
	"testing"

	"gitlab.com/uniget-org/cli/pkg/tool"
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
			expectOut: "-rwxr-xr-x bin/jq" + "\n" +
				"-rw-r--r-- share/man/man1/jq.1",
			outContains: true,
		},
	}

	runCobraTests(t, tt)
}
