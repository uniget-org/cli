package main

import (
	"github.com/spf13/cobra"
)

func initReinstallCmd() {
	rootCmd.AddCommand(reinstallCmd)
}

var reinstallCmd = &cobra.Command{
	Use:     "reinstall",
	Short:   "Reinstall tool",
	Long:    header + "\nReinstall tools",
	Args:    cobra.NoArgs,
	Run:     func(cmd *cobra.Command, args []string) {
		//
	},
}
