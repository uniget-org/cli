package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/tool"
	//"gitlab.com/uniget-org/cli/pkg/tool"
)

var (
	generateBaseImage   = "ubuntu:26.04"
	generateImageTarget = "usr/local"
	generatePinVersions = false
)

func initGenerateCmd() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&generateBaseImage, "base", generateBaseImage, "Base image to use")
	generateCmd.Flags().StringVar(&generateImageTarget, "root", generateImageTarget, "Root directory to install tools")
	generateCmd.Flags().BoolVar(&generatePinVersions, "pin-versions", generatePinVersions, "Pin tool versions (default: false)")
}

var generateCmd = &cobra.Command{
	Use: "generate",
	Aliases: []string{
		"g",
		"gen",
	},
	Short:   "Generate Dockerfile",
	Long:    constants.Header + "\nGenerate Dockerfile for a tool",
	Hidden:  true,
	GroupID: "helper",
	Args:    cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var requestedTools tool.Tools
		var plannedTools tool.Tools
		for _, toolName := range args {
			tool, err := tools.GetByName(toolName)
			if err != nil {
				return fmt.Errorf("unable to find tool %s: %s", toolName, err)
			}
			requestedTools.Tools = append(requestedTools.Tools, *tool)
		}
		for _, tool := range requestedTools.Tools {
			err := tools.ResolveDependencies(&plannedTools, tool.Name)
			if err != nil {
				return fmt.Errorf("unable to resolve dependencies for %s: %s", tool.Name, err)
			}
		}

		var result []string
		result = append(result, "#syntax=docker/dockerfile:1")
		result = append(result, "")
		for _, tool := range plannedTools.Tools {
			var toolVersion = "latest"
			if generatePinVersions {
				toolVersion = tool.Version
			}
			result = append(
				result,
				fmt.Sprintf("FROM %s%s:%s AS %s",
					constants.RegistryImagePrefix,
					tool.Name,
					toolVersion,
					tool.Name,
				),
			)
		}
		result = append(result, "")
		result = append(result, fmt.Sprintf("FROM %s", generateBaseImage))
		for _, tool := range plannedTools.Tools {
			result = append(
				result,
				fmt.Sprintf("COPY --link --from=%s%s:latest / /%s",
					constants.RegistryImagePrefix,
					tool.Name,
					generateImageTarget,
				),
			)
		}

		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "%s", strings.Join(result, "\n"))

		return nil
	},
}
