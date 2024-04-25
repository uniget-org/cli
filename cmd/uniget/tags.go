package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("update") {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		assertMetadataFileExists()
		assertMetadataIsLoaded()

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
		t.Style().Options.DrawBorder = false
		t.Style().Options.SeparateColumns = false
		t.Style().Options.SeparateFooter = false
		t.Style().Options.SeparateHeader = false
		t.Style().Options.SeparateRows = false

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
