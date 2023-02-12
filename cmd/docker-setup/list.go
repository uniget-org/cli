package main

import (
	"github.com/spf13/cobra"
)

func initListCmd() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List tools",
	Long:    header + "\nList tools",
	Args:    cobra.NoArgs,
	Run:     func(cmd *cobra.Command, args []string) {
		tools.List()
	},
}
