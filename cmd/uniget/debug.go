package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initDebugCmd() {
	rootCmd.AddCommand(debugCmd)
}

var debugCmd = &cobra.Command{
	Use:     "debug",
	Aliases: []string{},
	Short:   "Debug parameters",
	Long:    header + "\nDebug parameters",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "prefix: %s\n", viper.GetString("prefix"))
		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "target: %s\n", viper.GetString("target"))
		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "cacheRoot: %s\n", cacheRoot)
		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "cacheDirectory: %s\n", cacheDirectory)
		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "libRoot: %s\n", libRoot)
		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "libDirectory: %s\n", libDirectory)
		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "metadataFile: %s\n", metadataFile)

		for _, key := range viper.AllKeys() {
			//nolint:errcheck
			fmt.Fprintf(cmd.OutOrStdout(), "viper key: %s, value: %v\n", key, viper.Get(key))
		}

		return nil
	},
}
