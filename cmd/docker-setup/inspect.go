package main

import (
	"fmt"

	"github.com/spf13/cobra"
	//log "github.com/sirupsen/logrus"
	//"github.com/fatih/color"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

func initInspectCmd() {
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:     "inspect",
	Short:   "Inspect tool",
	Long:    header + "\nInspect tools",
	Args:    cobra.ExactArgs(1),
	RunE:    func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		tools, err := tool.LoadFromFile(metadataFile)
		if err != nil {
			return fmt.Errorf("Failed to load metadata from file %s: %s\n", metadataFile, err)
		}

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("Error getting tool %s\n", args[0])
		}
		tool.ReplaceVariables(prefix + target, arch, alt_arch)

		fmt.Printf("%s Inspecting %s %s\n", emoji_tool, tool.Name, tool.Version)
		err = tool.Inspect(registryImagePrefix, alt_arch)
		if err != nil {
			return fmt.Errorf("Unable to inspect %s: %s", tool, err)
		}

		return nil
	},
}
