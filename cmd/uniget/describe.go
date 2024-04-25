package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var describeOutput string

func initDescribeCmd() {
	describeCmd.Flags().StringVarP(&describeOutput, "output", "o", "pretty", "Output options: pretty, json, yaml")

	rootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:     "describe",
	Aliases: []string{"d", "info"},
	Short:   "Show detailed information about tools",
	Long:    header + "\nShow detailed information about tools",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("update") {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		toolName := args[0]
		tool, err := tools.GetByName(toolName)
		if err != nil {
			return fmt.Errorf("error getting tool %s", toolName)
		}
		tool.ReplaceVariables(viper.GetString("prefix")+"/"+viper.GetString("target"), arch, altArch)
		err = tool.GetMarkerFileStatus(viper.GetString("prefix") + "/" + cacheDirectory)
		if err != nil {
			return fmt.Errorf("error getting marker file status: %s", err)
		}
		err = tool.GetBinaryStatus()
		if err != nil {
			return fmt.Errorf("error getting binary status: %s", err)
		}
		err = tool.GetVersionStatus()
		if err != nil {
			return fmt.Errorf("error getting version status: %s", err)
		}

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

		return nil
	},
}
