package main

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("failed to get tool: %s", err)
		}
		tool.ReplaceVariables(prefix+"/"+target, arch, altArch)
		tool.GetMarkerFileStatus(prefix + "/" + cacheDirectory)
		tool.GetBinaryStatus()
		tool.GetVersionStatus()

		if !tool.Status.MarkerFilePresent && !tool.Status.BinaryPresent {
			pterm.Warning.Printfln("Tool %s is not installed", tool.Name)
			return fmt.Errorf("tool %s is not installed", tool.Name)
		}

		tool.ReplaceVariables(prefix+"/"+target, arch, altArch)
		version, err := tool.RunVersionCheck()
		if err != nil {
			return fmt.Errorf("failed to get version: %s", err)
		}
		fmt.Println(version)

		return nil
	},
}
