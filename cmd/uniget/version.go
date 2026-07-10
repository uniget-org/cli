package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

func initVersionCmd() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use: "version",
	Aliases: []string{
		"v",
		"ver",
	},
	Short: "Show version of installed tool",
	Long:  constants.Header + "\nShow version of installed tool",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if configuration.AutoUpdate {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("failed to get tool: %s", err)
		}

		err = tool.UpdateStatus(configuration.Prefix, configuration.Target, configuration.GetCacheDirectory(), configuration.Arch, configuration.AltArch)
		if err != nil {
			return fmt.Errorf("failed to update status for tool %s: %s", tool.Name, err)
		}

		markerFilePresent := fileExists(configuration.Prefix + "/" + configuration.GetLibDirectory() + "/manifests/" + tool.Name + ".txt")
		if !tool.Status.MarkerFilePresent && !tool.Status.BinaryPresent && !markerFilePresent {
			logging.Warning.Printfln("Tool %s is not installed", tool.Name)
			return fmt.Errorf("tool %s is not installed", tool.Name)
		}

		if tool.Check == "" {
			logging.Warning.Printfln("Tool %s does not support version check", tool.Name)
			fmt.Println(tool.Version)
			return nil
		}

		tool.ReplaceVariables(configuration.Prefix+"/"+configuration.Target, configuration.Arch, configuration.AltArch)
		version, err := tool.RunVersionCheck()
		if err != nil {
			return fmt.Errorf("failed to get version: %s", err)
		}
		fmt.Println(version)

		return nil
	},
}
