package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
)

func initDebugCmd() {
	rootCmd.AddCommand(debugCmd)
}

var debugCmd = &cobra.Command{
	Use:     "debug",
	Aliases: []string{},
	Short:   "Debug parameters",
	Long:    constants.Header + "\nDebug parameters",
	Hidden:  true,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "configuration: %s\n", configuration)

		info, ok := debug.ReadBuildInfo()
		if !ok {
			return fmt.Errorf("cannot read build info")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "build info: %+v\n", info)

		return nil
	},
}
