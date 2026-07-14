package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
)

var upgradeDryRun = false

func initUpgradeCmd() {
	upgradeCmd.Flags().BoolVar(&upgradeDryRun, "dry-run", upgradeDryRun, "Show tool(s) planned for upgrade")

	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Aliases: []string{},
	Short:   "Upgrade all tools",
	Long:    constants.Header + "\nUpgrade all tools to latest version",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		requestedTools, err := findInstalledTools(tools)
		if err != nil {
			return fmt.Errorf("failed to find installed tools: %s", err)
		}

		err = installTools(cmd.OutOrStdout(), requestedTools, false, upgradeDryRun, false, false, false)
		if err != nil {
			return fmt.Errorf("failed to upgrade tools: %s", err)
		}

		return nil
	},
}
