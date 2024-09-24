package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initEnvCmd() {
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"e"},
	Short:   "Display installation paths as environment variables",
	Long:    header + "\nDisplay installation paths as environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "UNIGET_PREFIX=%s\n", viper.GetString("prefix"))
		fmt.Fprintf(cmd.OutOrStdout(), "UNIGET_TARGET=%s/%s\n", viper.GetString("prefix"), viper.GetString("target"))
		fmt.Fprintf(cmd.OutOrStdout(), "UNIGET_CACHE_ROOT=%s/%s\n", viper.GetString("prefix"), cacheRoot)
		fmt.Fprintf(cmd.OutOrStdout(), "UNIGET_CACHE_DIRECTORY=%s/%s\n", viper.GetString("prefix"), cacheDirectory)
		fmt.Fprintf(cmd.OutOrStdout(), "UNIGET_LIB_ROOT=%s/%s\n", viper.GetString("prefix"), libRoot)
		fmt.Fprintf(cmd.OutOrStdout(), "UNIGET_LIB_DIRECTORY=%s/%s\n", viper.GetString("prefix"), libDirectory)
		fmt.Fprintf(cmd.OutOrStdout(), "UNIGET_METADATA_FILE=%s/%s\n", viper.GetString("prefix"), metadataFile)

		return nil
	},
}
