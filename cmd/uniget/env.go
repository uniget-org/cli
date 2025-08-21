package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func initEnvCmd() {
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"e"},
	Short:   "Display installation paths as environment variables",
	Long:    header + "\nDisplay installation paths as environment variables",
	Hidden:  true,
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
