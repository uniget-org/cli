package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
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
	Short:   "Show version of installed tool",
	Long:    constants.Header + "\nShow version of installed tool",
	GroupID: "tool",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		tools, err = configuration.EnsureMetadata()
		if err != nil {
			return handleError(err, "Failed to load metadata")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		tool, err := tools.GetByName(args[0])
		if err != nil {
			return handleError(err, "Unable to find tool %s", args[0])
		}

		err = tool.UpdateStatus(
			configuration.Prefix,
			configuration.Target,
			configuration.GetCacheDirectory(),
			configuration.Arch,
			configuration.AltArch,
		)
		if err != nil {
			return handleError(err, "Failed to update status for tool %s", tool.Name)
		}

		markerFilePresent := myos.FileExists(configuration.Prefix + "/" + configuration.GetLibDirectory() + "/manifests/" + tool.Name + ".txt")
		if !tool.Status.MarkerFilePresent && !tool.Status.BinaryPresent && !markerFilePresent {
			logging.Warning.Printfln("Tool %s is not installed", tool.Name)
			return handleError(errors.New("tool not installed"), "Tool %s is not installed", tool.Name)
		}

		if tool.Check == "" {
			logging.Warning.Printfln("Tool %s does not support version check", tool.Name)
			fmt.Println(tool.Version)
			return nil
		}

		tool.ReplaceVariables(configuration.Prefix+"/"+configuration.Target, configuration.Arch, configuration.AltArch)
		version, err := tool.RunVersionCheck()
		if err != nil {
			return handleError(err, "Failed to get version for tool %s", tool.Name)
		}
		fmt.Println(version)

		return nil
	},
}
