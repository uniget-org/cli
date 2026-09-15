package main

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

func initDepsCmd() {
	depsCmd.AddCommand(depsTreeCmd)
	depsCmd.AddCommand(depsWhoUsesCmd)
	rootCmd.AddCommand(depsCmd)
}

var depsCmd = &cobra.Command{
	Use: "dependencies",
	Aliases: []string{
		"deps",
	},
	Short:   "Show dependencies for a tool",
	Long:    constants.Header + "\nShow the dependencies for a tool",
	GroupID: "tool",
}

var depsTreeCmd = &cobra.Command{
	Use: "tree",
	Aliases: []string{
		"t",
	},
	Short: "Show dependency tree",
	Long:  constants.Header + "\nShow the dependency tree for a tool",
	Args:  cobra.OnlyValidArgs,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		tools, err = configuration.EnsureMetadata()
		if err != nil {
			return fmt.Errorf("error ensuring metadata: %s", err)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) > 0 {
			for _, toolName := range args {
				tool, err := tools.GetByName(toolName)
				if err != nil {
					return fmt.Errorf("error getting tool %s: %s", toolName, err)
				}

				showDependencyTree(tool, "")
			}

		} else {
			for _, tool := range tools.Tools {
				showDependencyTree(&tool, "")
			}
		}

		return nil
	},
}

func showDependencyTree(tool *tool.Tool, indent string) {
	fmt.Println(indent + tool.Name)
	for _, depName := range tool.RuntimeDependencies {
		dep, err := tools.GetByName(depName)
		if err != nil {
			fmt.Printf("error getting dependency %s: %s\n", depName, err)
			continue
		}

		showDependencyTree(dep, indent+"  ")
	}
}

var depsWhoUsesCmd = &cobra.Command{
	Use: "who-uses",
	Aliases: []string{
		"w",
	},
	Short: "Show who uses a tool",
	Long:  constants.Header + "\nShow which tools depend on a specific tool",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		tools, err = configuration.EnsureMetadata()
		if err != nil {
			return fmt.Errorf("error ensuring metadata: %s", err)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		toolName := args[0]

		// Show which tools depend on the selected tool
		for _, tool := range tools.Tools {
			if slices.Contains(tool.RuntimeDependencies, toolName) {
				fmt.Println(tool.Name)
			}
		}

		return nil
	},
}
