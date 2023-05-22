package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

func initTagsCmd() {
	rootCmd.AddCommand(tagsCmd)
}

var tagsCmd = &cobra.Command{
	Use:     "tags",
	Aliases: []string{"t"},
	Short:   "List tags",
	Long:    header + "\nList tags",
	Args:    cobra.NoArgs,
	RunE:    func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		tools, err := tool.LoadFromFile(metadataFileName)
		if err != nil {
			return fmt.Errorf("Failed to load metadata from file %s: %s\n", metadataFileName, err)
		}

		tags := make(map[string]int)
		for _, tool := range tools.Tools {
			for _, name := range tool.Tags {
				_, exists := tags[name]
				if !exists {
					tags[name] = 0
				}
				tags[name]++
			}
		}

		keys := make([]string, 0, len(tags))
		for key := range tags {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)

		t.AppendHeader(table.Row{"#", "Name", "# Tools"})

		for index, key := range keys {
			t.AppendRows([]table.Row{
				{index + 1, key, tags[key]},
			})
		}

		t.Render()

		return nil
	},
}
