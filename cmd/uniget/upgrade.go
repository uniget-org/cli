package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

func initUpgradeCmd() {
	upgradeCmd.Flags().BoolVar(&dryRun, "plan", false, "Show tool(s) planned installation")
	upgradeCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show tool(s) planned for installation")
	upgradeCmd.MarkFlagsMutuallyExclusive("plan", "dry-run")
	err := upgradeCmd.Flags().MarkHidden("plan")
	if err != nil {
		logging.Error.Printfln("Unable to mark plan flag as hidden: %s", err)
	}

	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Aliases: []string{},
	Short:   "Upgrade all tools",
	Long:    header + "\nUpgrade all tools to latest version",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("update") {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		requestdTools, err := findInstalledTools(tools)
		if err != nil {
			return fmt.Errorf("failed to find installed tools: %s", err)
		}

		err = installTools(cmd.OutOrStdout(), requestdTools, false, dryRun, false, false, false)
		if err != nil {
			return fmt.Errorf("failed to install tools: %s", err)
		}

		return nil
	},
}
