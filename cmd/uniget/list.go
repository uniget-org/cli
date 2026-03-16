package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
	"gopkg.in/yaml.v3"
)

var installedOnly bool
var upgradableOnly bool
var listOutput string

func initListCmd() {
	listCmd.Flags().BoolVar(&installedOnly, "installed", false, "List only installed tools")
	listCmd.Flags().BoolVar(&upgradableOnly, "upgradable", false, "List only upgradable tools")
	listCmd.Flags().StringVarP(&listOutput, "output", "o", "pretty", "Output options: pretty, json, yaml")

	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use: "list",
	Aliases: []string{
		"l",
		"get",
	},
	Short: "List tools",
	Long:  header + "\nList tools",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("update") {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		var listTools tool.Tools

		if installedOnly {
			var installedTools tool.Tools
			for index := range tools.Tools {
				checkClientVersionRequirement(&tools.Tools[index])

				err := tools.Tools[index].UpdateStatus(viper.GetString("prefix"), viper.GetString("target"), cacheDirectory, arch, altArch)
				if err != nil {
					return fmt.Errorf("failed to update status for tool %s: %s", tools.Tools[index].Name, err)
				}

				if tools.Tools[index].IsInstalled() {
					installedTools.Tools = append(installedTools.Tools, tools.Tools[index])
				}
			}
			listTools = installedTools

		} else if upgradableOnly {
			var installedTools tool.Tools
			for index := range tools.Tools {
				checkClientVersionRequirement(&tools.Tools[index])

				err := tools.Tools[index].UpdateStatus(viper.GetString("prefix"), viper.GetString("target"), cacheDirectory, arch, altArch)
				if err != nil {
					return fmt.Errorf("failed to update status for tool %s: %s", tools.Tools[index].Name, err)
				}

				if tools.Tools[index].IsUpgradable() {
					installedTools.Tools = append(installedTools.Tools, tools.Tools[index])
				}
			}
			listTools = installedTools

		} else {
			listTools = tools
		}

		switch listOutput {
		case "pretty":
			listTools.List(cmd.OutOrStdout())
		case "json":
			data, err := json.Marshal(listTools)
			if err != nil {
				return fmt.Errorf("failed to marshal to json: %s", err)
			}
			fmt.Println(string(data))
		case "yaml":
			yamlEncoder := yaml.NewEncoder(cmd.OutOrStdout())
			yamlEncoder.SetIndent(2)
			defer func() {
				err := yamlEncoder.Close()
				if err != nil {
					logging.Warning.Printfln("failed to close yaml encoder: %s", err)
				}
			}()
			err := yamlEncoder.Encode(listTools)
			if err != nil {
				return fmt.Errorf("failed to encode yaml: %s", err)
			}
		default:
			return fmt.Errorf("invalid output format: %s", listOutput)
		}

		return nil
	},
}
