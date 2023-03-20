package main

import (
	"github.com/spf13/cobra"
)

func initInspectCmd() {
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:     "inspect",
	Short:   "Inspect tool",
	Long:    header + "\nInspect tools",
	Args:    cobra.NoArgs,
	Run:     func(cmd *cobra.Command, args []string) {
		//
	},
}
