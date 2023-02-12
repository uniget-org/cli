package main

import (
	"os"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func initTagsCmd() {
	rootCmd.AddCommand(tagsCmd)
}

var tagsCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "List tools and tags",
	Long:    header + "\nList tools and tags",
	Args:    cobra.NoArgs,
	RunE:    func(cmd *cobra.Command, args []string) error {
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
