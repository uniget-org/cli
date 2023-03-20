package main

import (
	"github.com/spf13/cobra"
)

func initUpdateCmd() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update tool",
	Long:    header + "\nUpdate tools",
	Args:    cobra.NoArgs,
	Run:     func(cmd *cobra.Command, args []string) {
		//
	},
}
