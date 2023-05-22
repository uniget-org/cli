package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

func initListCmd() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "get"},
	Short:   "List tools",
	Long:    header + "\nList tools",
	Args:    cobra.NoArgs,
	RunE:    func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		tools, err := tool.LoadFromFile(metadataFileName)
		if err != nil {
			return fmt.Errorf("Failed to load metadata from file %s: %s\n", metadataFileName, err)
		}

		tools.List()

		return nil
	},
}
