package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func initDebugCmd() {
	rootCmd.AddCommand(debugCmd)
}

var debugCmd = &cobra.Command{
	Use:     "debug",
	Aliases: []string{},
	Short:   "Debug parameters",
	Long:    header + "\nDebug parameters",
	Hidden:  true,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "config: %+v\n", config)

		return nil
	},
}
