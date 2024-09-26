package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/uniget-org/cli/pkg/cache"
	"github.com/uniget-org/cli/pkg/logging"
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

	toolCache = cache.NewNoneCache()

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
