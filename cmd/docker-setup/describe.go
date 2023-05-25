package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var describeOutput string

func initDescribeCmd() {
	rootCmd.AddCommand(describeCmd)

	describeCmd.Flags().StringVarP(&describeOutput, "output", "o", "pretty", "Output options: pretty, json, yaml")
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

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("error getting tool %s", args[0])
		}
		tool.ReplaceVariables(prefix+target, arch, alt_arch)

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
		}

		return nil
	},
}
