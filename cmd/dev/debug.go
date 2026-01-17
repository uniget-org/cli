package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func initDebugCmd() {
	rootCmd.AddCommand(debugCmd)
}

var debugCmd = &cobra.Command{
	Use: "debug",
	Aliases: []string{
		"d",
	},
	Short: "Build tool",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return unigetToolsNames, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		toolName := args[0]

		if !unigetTools.Exists(toolName) {
			return fmt.Errorf("tool %s does not exist", toolName)
		}
		tool := unigetTools.Tools[toolName]

		var err error

		err = buildLocal(cmd, tool)
		if err != nil {
			return fmt.Errorf("error building tool locally: %w", err)
		}

		err = runLocal(cmd, tool)
		if err != nil {
			return fmt.Errorf("error running tool locally: %w", err)
		}

		return nil
	},
}

func buildLocal(cmd *cobra.Command, tool UnigetTool) error {
	fmt.Printf("Building tool: %s\n", tool.Tool.Name)
	fmt.Printf("  Version: %s\n", tool.Tool.Version)
	fmt.Printf("  Deps: %v\n", strings.Join(tool.Tool.RuntimeDependencies, ","))
	fmt.Printf("  Tags: %v\n", strings.Join(tool.Tool.Tags, ","))
	fmt.Println()

	err := runCommand(cmd,
		[]string{
			"docker",
			"buildx",
			"--builder=default",
			"debug",
			"--on=error",
			"--invoke=/bin/bash",
			"build",
			fmt.Sprintf("%s/%s", unigetTools.Directory, tool.Subdirectory),
			fmt.Sprintf("--build-arg=branch=%s", dockerTag),
			fmt.Sprintf("--build-arg=ref=%s", dockerTag),
			fmt.Sprintf("--build-arg=name=%s", tool.Tool.Name),
			fmt.Sprintf("--build-arg=version=%s", tool.Tool.Version),
			fmt.Sprintf("--build-arg=deps=%s", strings.Join(tool.Tool.RuntimeDependencies, ",")),
			fmt.Sprintf("--build-arg=tags=%s", strings.Join(tool.Tool.Tags, ",")),
			fmt.Sprintf("--cache-from=%s/%s%s:latest", registryHost, repositoryPrefix, tool.Tool.Name),
			"--platform=linux/amd64",
			fmt.Sprintf("--tag=%s/%s%s:%s", registryHost, repositoryPrefix, tool.Tool.Name, tool.Tool.Version),
			"--target=prepare",
			"--output=type=docker,oci-mediatypes=true",
			"--progress=plain",
		},
		[]string{
			"BUILDX_EXPERIMENTAL=1",
		},
	)
	if err != nil {
		return fmt.Errorf("error building tool locally: %w", err)
	}

	return nil
}

func runLocal(cmd *cobra.Command, tool UnigetTool) error {
	err := runCommand(cmd,
		[]string{
			"docker",
			"container",
			"run",
			"--interactive",
			"--tty",
			"--privileged",
			fmt.Sprintf("--env=name=%s", tool.Tool.Name),
			fmt.Sprintf("--env=version=%s", tool.Tool.Version),
			"--rm",
			fmt.Sprintf("%s/%s%s:%s", registryHost, repositoryPrefix, tool.Tool.Name, tool.Tool.Version),
			"bash",
			"--login",
			"+o",
			"errexit",
			"+o",
			"pipefail",
		},
		[]string{},
	)
	if err != nil {
		fmt.Println("could not run command: ", err)
	}

	return nil
}
