package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var baseImage string

func initGenerateCmd() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&baseImage, "base", "ubuntu:22.04", "Base image to use")
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Dockerfile",
	Long:  header + "\nGenerate Dockerfile for a tool",
	Args:  cobra.MinimumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var result []string

		result = append(result, fmt.Sprintf("FROM %s", baseImage))
		for _, toolName := range args {
			tool, err := tools.GetByName(toolName)
			if err != nil {
				return fmt.Errorf("failed to get tool %s: %w", toolName, err)
			}
			result = append(result, fmt.Sprintf("COPY --link --from=%s%s:%s / /", registryImagePrefix, tool.Name, strings.Replace(tool.Version, "+", "-", -1)))
		}
		fmt.Printf("%s", strings.Join(result, "\n"))

		return nil
	},
}
