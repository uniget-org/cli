package main

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

var describeOutput string

func initDescribeCmd() {
	rootCmd.AddCommand(describeCmd)

	describeCmd.Flags().StringVarP(&describeOutput, "output", "o", "pretty", "Output options: pretty, json, yaml")
}

var describeCmd = &cobra.Command{
	Use:     "describe",
	Aliases: []string{"d"},
	Short:   "Show detailed information about tools",
	Long:    header + "\nShow detailed information about tools",
	Args:    cobra.ExactArgs(1),
	RunE:    func(cmd *cobra.Command, args []string) error {
		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("Error getting tool %s\n", args[0])
		}

		if describeOutput == "pretty" {
			tool.Print()

		} else if describeOutput == "json" {
			data, _ := json.Marshal(tool)
			fmt.Println(string(data))

		} else if describeOutput == "yaml" {
			data, _ := yaml.Marshal(tool)
			fmt.Println(string(data))
		}

		return nil
	},
}
