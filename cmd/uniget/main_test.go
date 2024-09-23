package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type cobraTestOpts struct {
	stdin io.Reader
}

func cobraTest(t *testing.T, opts *cobraTestOpts, args ...string) (string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	if opts != nil && opts.stdin != nil {
		rootCmd.SetIn(opts.stdin)
	}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	return strings.TrimSpace(buf.String()), err
}
