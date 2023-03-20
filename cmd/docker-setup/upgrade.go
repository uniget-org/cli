package main

import (
	"github.com/spf13/cobra"
)

func initUpgradeCmd() {
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Short:   "Upgrade tool",
	Long:    header + "\nUpgrade tools",
	Args:    cobra.NoArgs,
	Run:     func(cmd *cobra.Command, args []string) {
		//
	},
}
