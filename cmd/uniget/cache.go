package main

import (
	"github.com/spf13/cobra"

	"gitlab.com/uniget-org/cli/internal/constants"
)

func initCacheCmd() {
	rootCmd.AddCommand(cacheCmd)

	cacheCmd.AddCommand(cacheInfoCmd)
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cachePruneCmd)
}

var cacheCmd = &cobra.Command{
	Use: "cache",
	Aliases: []string{
		"c",
	},
	Short:   "Manage the cache",
	Long:    constants.Header + "\nManage the cache",
	GroupID: "config",
	Args:    cobra.NoArgs,
}

var cacheInfoCmd = &cobra.Command{
	Use: "info",
	Aliases: []string{
		"i",
	},
	Short: "Display cache information (no-op)",
	Long:  constants.Header + "\nDisplay cache information (no-op)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var cacheStatsCmd = &cobra.Command{
	Use: "stats",
	Aliases: []string{
		"s",
	},
	Short: "Display cache statistics (no-op)",
	Long:  constants.Header + "\nDisplay cache statistics (no-op)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var cacheListCmd = &cobra.Command{
	Use: "list",
	Aliases: []string{
		"l",
		"ls",
	},
	Short: "Display cache contents (no-op)",
	Long:  constants.Header + "\nDisplay cache contents (no-op)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var cachePruneCmd = &cobra.Command{
	Use: "prune",
	Aliases: []string{
		"p",
	},
	Short: "Remove unused cache entries (no-op)",
	Long:  constants.Header + "\nRemove unused cache entries (no-op)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
