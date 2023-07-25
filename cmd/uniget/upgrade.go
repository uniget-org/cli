package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func initUpgradeCmd() {
	upgradeCmd.Flags().BoolVar(&plan, "plan", false, "Show tool(s) planned installation")

	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade all tools",
	Long:  header + "\nUpgrade all tools to latest version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		requestdTools, err := findInstalledTools(tools)
		if err != nil {
			return fmt.Errorf("failed to find installed tools: %s", err)
		}

		installTools(requestdTools, false, plan, false, false, false)

		return nil
	},
}
