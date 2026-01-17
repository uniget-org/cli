package main

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func runCommand(cmd *cobra.Command, command []string, environment []string) error {
	// #nosec G204 -- Not exported and only used internally
	execCmd := exec.Command(command[0], command[1:]...)
	execCmd.Env = append(os.Environ(), environment...)
	execCmd.Stdin = cmd.InOrStdin()
	execCmd.Stdout = cmd.OutOrStdout()
	execCmd.Stderr = cmd.ErrOrStderr()
	return execCmd.Run()
}
