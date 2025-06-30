package main

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func initEditCmd() {
	rootCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:       "edit",
	Aliases:   []string{"e", "ed"},
	Short:     "Edit tool in Visual Studio Code",
	ValidArgs: unigetToolsNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, toolName := range args {
			if !unigetTools.Exists(toolName) {
				return fmt.Errorf("tool %s does not exist", toolName)
			}

			fmt.Printf("Editing tool: %s\n", toolName)
			tool := unigetTools.Tools[toolName]
			cmd := exec.Command(
				"code",
				"--goto",
				fmt.Sprintf("%s/%s/manifest.yaml", unigetTools.Directory, tool.Subdirectory),
			)
			_, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error opening tool in Visual Studio Code: %w", err)
			}
		}
		return nil
	},
}
