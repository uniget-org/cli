package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func runCommand(cmd *cobra.Command, command []string, environment []string) error {
	execCmd := exec.Command(command[0], command[1:]...)
	execCmd.Env = append(os.Environ(), environment...)
	execCmd.Stdin = cmd.InOrStdin()
	execCmd.Stdout = cmd.OutOrStdout()
	execCmd.Stderr = cmd.ErrOrStderr()
	return execCmd.Run()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return out.Sync()
}
