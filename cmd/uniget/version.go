package main

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initVersionCmd() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show version of installed tool",
	Long:    header + "\nShow version of installed tool",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("failed to get tool: %s", err)
		}
		tool.ReplaceVariables(viper.GetString("prefix")+"/"+viper.GetString("target"), arch, altArch)
		err = tool.GetMarkerFileStatus(viper.GetString("prefix") + "/" + cacheDirectory)
		if err != nil {
			return fmt.Errorf("failed to get marker file status: %s", err)
		}
		err = tool.GetBinaryStatus()
		if err != nil {
			return fmt.Errorf("failed to get binary status: %s", err)
		}
		err = tool.GetVersionStatus()
		if err != nil {
			return fmt.Errorf("failed to get version status: %s", err)
		}

		markerFilePresent := false
		if fileExists(viper.GetString("prefix") + "/" + libDirectory + "/manifests/" + tool.Name + ".txt") {
			markerFilePresent = true
		}

		if !tool.Status.MarkerFilePresent && !tool.Status.BinaryPresent && !markerFilePresent {
			pterm.Warning.Printfln("Tool %s is not installed", tool.Name)
			return fmt.Errorf("tool %s is not installed", tool.Name)
		}

		if tool.Check == "" {
			pterm.Warning.Printfln("Tool %s does not support version check", tool.Name)
			fmt.Println(tool.Version)
			return nil
		}

		tool.ReplaceVariables(viper.GetString("prefix")+"/"+viper.GetString("target"), arch, altArch)
		version, err := tool.RunVersionCheck()
		if err != nil {
			return fmt.Errorf("failed to get version: %s", err)
		}
		fmt.Println(version)

		return nil
	},
}
