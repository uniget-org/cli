package main

import (
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
	Short:   "Upgrade tools",
	Long:    constants.Header + "\nUpgrade tools to latest version",
	GroupID: "tool",
	Args:    cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		tools, err = configuration.EnsureMetadata()
		if err != nil {
			return handleError(err, "Failed to load metadata")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		requestedTools, err := findInstalledTools(tools)
		if err != nil {
			return handleError(err, "Failed to find installed tools")
		}

		return installTools(cmd.OutOrStdout(), requestedTools, false, upgradeDryRun, false, false, false)
	},
}
