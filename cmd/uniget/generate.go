package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/tool"
	//"github.com/uniget-org/cli/pkg/tool"
)

var (
	baseImage   = "ubuntu:24.04"
	imageTarget = "usr/local"
	pinVersions = false
)

func initGenerateCmd() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&baseImage, "base", baseImage, "Base image to use")
	generateCmd.Flags().StringVar(&imageTarget, "root", imageTarget, "Root directory to install tools")
	generateCmd.Flags().BoolVar(&pinVersions, "pin-versions", pinVersions, "Pin tool versions (default: false)")
}

var generateCmd = &cobra.Command{
	Use:    "generate",
	Short:  "Generate Dockerfile",
	Long:   header + "\nGenerate Dockerfile for a tool",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
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
			if pinVersions {
				toolVersion = tool.Version
			}
			result = append(result, fmt.Sprintf("FROM %s%s:%s AS %s", registryImagePrefix, tool.Name, toolVersion, tool.Name))
		}
		result = append(result, "")
		result = append(result, fmt.Sprintf("FROM %s", baseImage))
		for _, tool := range plannedTools.Tools {
			result = append(result, fmt.Sprintf("COPY --link --from=%s%s:latest / /%s", registryImagePrefix, tool.Name, imageTarget))
		}

		//nolint:errcheck
		fmt.Fprintf(cmd.OutOrStdout(), "%s", strings.Join(result, "\n"))

		return nil
	},
}
