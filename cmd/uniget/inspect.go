package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uniget-org/cli/pkg/logging"

	"github.com/uniget-org/cli/pkg/tool"
)

var toolVersion string
var rawInspect bool

func initInspectCmd() {
	inspectCmd.Flags().StringVar(&toolVersion, "version", "", "Inspect a specific version of the tool")
	inspectCmd.Flags().BoolVar(&rawInspect, "raw", false, "Show raw contents")

	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect tool",
	Long:  header + "\nInspect tools",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var inspectTool *tool.Tool

		if len(toolVersion) == 0 {
			assertMetadataFileExists()
			assertMetadataIsLoaded()

			inspectTool, err = tools.GetByName(args[0])
			if err != nil {
				return fmt.Errorf("error getting tool %s", args[0])
			}
			inspectTool.ReplaceVariables(viper.GetString("prefix")+viper.GetString("target"), arch, altArch)

		} else {
			inspectTool = &tool.Tool{
				Name:    args[0],
				Version: toolVersion,
			}
		}

		logging.Info.Printfln("Inspecting %s %s\n", inspectTool.Name, inspectTool.Version)
		err = inspectTool.Inspect(registryImagePrefix, altArch, rawInspect)
		if err != nil {
			return fmt.Errorf("unable to inspect %s: %s", inspectTool.Name, err)
		}

		return nil
	},
}
