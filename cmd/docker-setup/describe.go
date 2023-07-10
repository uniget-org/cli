package main

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var describeOutput string

func initDescribeCmd() {
	describeCmd.Flags().StringVarP(&describeOutput, "output", "o", "pretty", "Output options: pretty, json, yaml")

	rootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:       "describe",
	Aliases:   []string{"d", "info"},
	Short:     "Show detailed information about tools",
	Long:      header + "\nShow detailed information about tools",
	Args:      cobra.ExactArgs(1),
	ValidArgs: tools.GetNames(),
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		toolName := args[0]
		tool, err := tools.GetByName(toolName)
		if err != nil {
			return fmt.Errorf("error getting tool %s", toolName)
		}
		tool.ReplaceVariables(prefix+"/"+target, arch, altArch)
		tool.GetMarkerFileStatus(prefix + "/" + cacheDirectory)
		tool.GetBinaryStatus()
		tool.GetVersionStatus()

		if describeOutput == "pretty" {
			tool.Print()

		} else if describeOutput == "json" {
			data, err := json.Marshal(tool)
			if err != nil {
				return fmt.Errorf("failed to marshal to json: %s", err)
			}
			fmt.Println(string(data))

		} else if describeOutput == "yaml" {
			yamlEncoder := yaml.NewEncoder(os.Stdout)
			yamlEncoder.SetIndent(2)
			defer yamlEncoder.Close()
			err := yamlEncoder.Encode(tool)
			if err != nil {
				return fmt.Errorf("failed to encode yaml: %s", err)
			}

		} else {
			return fmt.Errorf("invalid output format: %s", describeOutput)
		}

		if noInteractive || !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
			return nil
		}

		fmt.Println()
		primaryOptions := []string{"Abort", "Plan", "Install", "Uninstall"}
		printer := pterm.DefaultInteractiveSelect.WithOptions(primaryOptions)
		printer.DefaultText = "What do you want to do?"
		selectedOption, _ := printer.Show()
		switch selectedOption {
		case "Abort":
			return nil
		case "Plan":
			err := installToolsByName([]string{toolName}, false, true, false, false, false)
			if err != nil {
				return err
			}
			continueWithInstall, _ := pterm.DefaultInteractiveConfirm.Show()
			if continueWithInstall {
				return installToolsByName([]string{toolName}, false, false, false, false, false)
			}
		case "Install":
			return installToolsByName([]string{toolName}, false, false, false, false, false)
		case "Uninstall":
			return uninstallTool(toolName)
		default:
			return fmt.Errorf("invalid option: %s", selectedOption)
		}

		return nil
	},
}
