package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initUpgradeCmd() {
	upgradeCmd.Flags().BoolVar(&plan, "plan", false, "Show tool(s) planned installation")

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

		err = installTools(cmd.OutOrStdout(), requestdTools, false, plan, false, false, false)
		if err != nil {
			return fmt.Errorf("failed to install tools: %s", err)
		}

		return nil
	},
}
