package main

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initHealthcheckCmd() {
	rootCmd.AddCommand(healthcheckCmd)
}

var healthcheckCmd = &cobra.Command{
	Use:     "healthcheck",
	Aliases: []string{"health"},
	Short:   "Check health of installed tool",
	Long:    header + "\nCheck health of installed tool",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		toolName := args[0]
		tool, err := tools.GetByName(toolName)
		if err != nil {
			return fmt.Errorf("error getting tool %s", toolName)
		}
		tool.ReplaceVariables(viper.GetString("prefix")+"/"+viper.GetString("target"), arch, altArch)
		err = tool.GetMarkerFileStatus(viper.GetString("prefix") + "/" + cacheDirectory)
		if err != nil {
			return fmt.Errorf("error getting marker file status: %s", err)
		}
		err = tool.GetBinaryStatus()
		if err != nil {
			return fmt.Errorf("error getting binary status: %s", err)
		}
		err = tool.GetVersionStatus()
		if err != nil {
			return fmt.Errorf("error getting version status: %s", err)
		}

		testFailed := false

		if tool.Status.MarkerFilePresent {
			pterm.Success.Printfln("%s: Marker file is present", tool.Name)
		} else {
			pterm.Warning.Printfln("%s: Marker file is not present", tool.Name)
		}
		if tool.Binary == "false" {
			pterm.Warning.Printfln("%s: Tool does not have a binary", tool.Name)
		} else if tool.Status.BinaryPresent {
			pterm.Success.Printfln("%s: Binary is present (%s)", tool.Name, tool.Binary)
		} else {
			pterm.Error.Printfln("%s: Binary is not present (%s)", tool.Name, tool.Binary)
			testFailed = true
		}

		markerFilePresent := false
		if fileExists(viper.GetString("prefix") + "/" + libDirectory + "/manifests/" + tool.Name + ".txt") {
			markerFilePresent = true
		}

		if !tool.Status.MarkerFilePresent && !tool.Status.BinaryPresent && !markerFilePresent {
			pterm.Warning.Printfln("Tool %s is not installed", tool.Name)
			testFailed = true
		}

		if tool.Check == "" {
			pterm.Warning.Printfln("%s: Tool does not support version check", tool.Name)
			pterm.Info.Printfln("%s: Manifest version is %s", tool.Name, tool.Version)

		} else {
			tool.ReplaceVariables(viper.GetString("prefix")+"/"+viper.GetString("target"), arch, altArch)
			version, err := tool.RunVersionCheck()
			if err != nil {
				pterm.Error.Printfln("%s: Error getting version: %s", tool.Name, err)
				testFailed = true
			}
			pterm.Success.Printfln("%s: Installed version is %s", tool.Name, version)
		}

		if testFailed {
			pterm.Error.Printfln("%s: Healthcheck failed", tool.Name)
			return fmt.Errorf("healthcheck failed for %s", tool.Name)
		}

		return nil
	},
}
