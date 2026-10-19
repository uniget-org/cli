package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"gitlab.com/uniget-org/cli/internal/config"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

type cobraTest struct {
	name        string
	args        []string
	expectErr   error
	expectOut   string
	outContains bool
}

type cobraTestOpts struct {
	stdin io.Reader
}

func runCobraTest(t *testing.T, opts *cobraTestOpts, args ...string) (string, error) {
	t.Helper()

	configuration = config.NewDefaultConfig()
	resetCommandFlags()

	buf := new(bytes.Buffer)
	if opts != nil && opts.stdin != nil {
		rootCmd.SetIn(opts.stdin)
	}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	logging.OutputWriter = rootCmd.OutOrStdout()
	logging.ErrorWriter = rootCmd.ErrOrStderr()
	logging.Init()

	err := rootCmd.Execute()
	return strings.TrimSpace(buf.String()), err
}

// runCobraTest also returns output which is written directly to stdout by
// commands using fmt.Print instead of the cobra writers.
func runCobraTestCaptureStdout(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	savedStdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = savedStdout
	}()

	out, cobraErr := runCobraTest(t, nil, args...)

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close pipe: %v", err)
	}
	stdout, readErr := io.ReadAll(reader)
	if readErr != nil {
		t.Fatalf("failed to read stdout: %v", readErr)
	}

	return out, strings.TrimSpace(string(stdout)), cobraErr
}

func runCobraTests(t *testing.T, tt []cobraTest) {
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runCobraTest(t, nil, tc.args...)

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
				t.Errorf("unexpected output, expected <%s>, received <%s>", tc.expectOut, out)
			}
		})
	}
}

// resetCommandFlags resets flags which are bound to global variables in
// package main. Cobra does not reset flag values or the Changed state
// between executions which leaks state between tests.
func resetCommandFlags() {
	resetFlag(searchCmd, "output", "table")
	for _, name := range []string{
		"only-names", "no-names",
		"only-description", "no-description",
		"only-tags", "no-tags",
		"only-deps", "no-deps",
	} {
		resetFlag(searchCmd, name, "false")
	}

	resetFlag(listCmd, "output", "pretty")
	resetFlag(listCmd, "installed", "false")
	resetFlag(listCmd, "upgradable", "false")

	resetFlag(describeCmd, "output", "pretty")
	resetFlag(describeCmd, "versions", "false")
	resetFlag(describeCmd, "upstream-versions", "false")

	resetFlag(messageCmd, "find", "false")
	resetFlag(messageCmd, "list", "false")

	resetFlag(manpagesCmd, "path", "share/man")
}

func resetFlag(cmd *cobra.Command, name string, value string) {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return
	}

	_ = flag.Value.Set(value)
	flag.Changed = false
}
