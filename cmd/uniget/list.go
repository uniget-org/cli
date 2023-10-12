package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/tool"
	"gopkg.in/yaml.v3"
)

var installedOnly bool
var listOutput string

func initListCmd() {
	listCmd.Flags().BoolVar(&installedOnly, "installed", false, "List only installed tools")
	listCmd.Flags().StringVarP(&listOutput, "output", "o", "pretty", "Output options: pretty, json, yaml")

	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "get"},
	Short:   "List tools",
	Long:    header + "\nList tools",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		var listTools tool.Tools

		if installedOnly {
			var installedTools tool.Tools
			for index := range tools.Tools {
				tools.Tools[index].ReplaceVariables(prefix+"/"+target, arch, altArch)
				err := tools.Tools[index].GetMarkerFileStatus(prefix + "/" + cacheDirectory)
				if err != nil {
					return fmt.Errorf("error getting marker file status: %s", err)
				}
				err = tools.Tools[index].GetBinaryStatus()
				if err != nil {
					return fmt.Errorf("error getting binary status: %s", err)
				}
				err = tools.Tools[index].GetVersionStatus()
				if err != nil {
					return fmt.Errorf("error getting version status: %s", err)
				}

				if tools.Tools[index].Status.VersionMatches {
					installedTools.Tools = append(installedTools.Tools, tools.Tools[index])
				}
			}
			listTools = installedTools

		} else {
			listTools = tools
		}

		if listOutput == "pretty" {
			listTools.List()

		} else if listOutput == "json" {
			data, err := json.Marshal(listTools)
			if err != nil {
				return fmt.Errorf("failed to marshal to json: %s", err)
			}
			fmt.Println(string(data))

		} else if listOutput == "yaml" {
			yamlEncoder := yaml.NewEncoder(os.Stdout)
			yamlEncoder.SetIndent(2)
			defer yamlEncoder.Close()
			err := yamlEncoder.Encode(listTools)
			if err != nil {
				return fmt.Errorf("failed to encode yaml: %s", err)
			}

		} else {
			return fmt.Errorf("invalid output format: %s", listOutput)
		}

		return nil
	},
}
