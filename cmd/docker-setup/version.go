package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func initVersionCmd() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show version",
	Long:    header + "\nShow version",
	Args:    cobra.NoArgs,
	Run:     func(cmd *cobra.Command, args []string) {
		fmt.Printf("docker-setup version %s\n", version)
	},
}
