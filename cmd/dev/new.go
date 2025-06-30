package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func initNewCmd() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:       "new",
	Aliases:   []string{"n", "create", "c"},
	Short:     "Create new tool",
	ValidArgs: unigetToolsNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for clean git working directory

		for _, toolName := range args {
			if unigetTools.Exists(toolName) {
				return fmt.Errorf("tool %s already exists", toolName)
			}

			fmt.Printf("Creating tool: %s\n", toolName)
			err := os.Mkdir(fmt.Sprintf("%s/%s", unigetTools.Directory, toolName), 0755)
			if err != nil {
				return fmt.Errorf("error creating tool directory: %w", err)
			}
			copyTemplates(toolName)
		}
		return nil
	},
}

func copyTemplates(toolName string) error {
	templateDir := unigetTools.BaseDirectory + "/@template"
	toolsDir := unigetTools.Directory

	files := []string{"manifest.yaml", "Dockerfile.template"}
	for _, file := range files {
		src := fmt.Sprintf("%s/%s", templateDir, file)
		dest := fmt.Sprintf("%s/%s/%s", toolsDir, toolName, file)
		if err := copyFile(src, dest); err != nil {
			return fmt.Errorf("error copying template %s: %w", file, err)
		}
	}
	return nil
}
