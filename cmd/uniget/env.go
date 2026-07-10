package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
)

func initEnvCmd() {
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use: "env",
	Aliases: []string{
		"e",
		"environment",
	},
	Short:  "Display installation paths as environment variables",
	Long:   constants.Header + "\nDisplay installation paths as environment variables",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "UNIGET_") {
				//nolint:errcheck
				fmt.Fprintf(cmd.OutOrStdout(), "env: %s\n", env)
			}
		}

		return nil
	},
}
