package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

var (
	searchOnlyInName        bool
	searchNotInName         bool
	searchOnlyInDescription bool
	searchNotInDescription  bool
	searchOnlyInTags        bool
	searchNotInTags         bool
	searchOnlyInDeps        bool
	searchNotInDeps         bool
	searchOutputFormat      string
)

func initSearchCmd() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolVar(&searchOnlyInName, "only-names", false, "Search only in names")
	searchCmd.Flags().BoolVar(&searchNotInName, "no-names", false, "Do not search in names")
	searchCmd.Flags().BoolVar(&searchOnlyInDescription, "only-description", false, "Search only in description")
	searchCmd.Flags().BoolVar(&searchNotInDescription, "no-description", false, "Do not search in description")
	searchCmd.Flags().BoolVar(&searchOnlyInTags, "only-tags", false, "Search only on tags")
	searchCmd.Flags().BoolVar(&searchNotInTags, "no-tags", false, "Do not search in tags")
	searchCmd.Flags().BoolVar(&searchOnlyInDeps, "only-deps", false, "Search only in dependencies")
	searchCmd.Flags().BoolVar(&searchNotInDeps, "no-deps", false, "Do not search in dependencies")
	searchCmd.Flags().StringVar(&searchOutputFormat, "output", "table", "Output format (table, name, json)")

	searchCmd.MarkFlagsMutuallyExclusive("only-names", "no-names")
	searchCmd.MarkFlagsMutuallyExclusive("only-description", "no-description")
	searchCmd.MarkFlagsMutuallyExclusive("only-tags", "no-tags")
	searchCmd.MarkFlagsMutuallyExclusive("only-deps", "no-deps")
}

var searchCmd = &cobra.Command{
	Use: "search",
	Aliases: []string{
		"s",
		"find",
		"f",
	},
	Short: "Search for tools",
	Long:  constants.Header + "\nSearch for tools",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if searchOutputFormat != "table" && searchOutputFormat != "name" && searchOutputFormat != "json" {
			return fmt.Errorf("error: output format %s not supported", searchOutputFormat)
		}

		if (searchOnlyInName && searchOnlyInTags) ||
			(searchOnlyInName && searchOnlyInDeps) ||
			(searchOnlyInName && searchNotInDescription) ||
			(searchOnlyInDescription && searchOnlyInTags) ||
			(searchOnlyInDescription && searchOnlyInDeps) ||
			(searchOnlyInTags && searchOnlyInDeps) {
			return fmt.Errorf("error: Can only process one of only-names, only-description, only-tags and only-deps at the same time")
		}

		results := tools
		for _, term := range args {
			results = results.Find(
				term,
				!searchNotInName && !searchOnlyInDescription && !searchOnlyInTags && !searchOnlyInDeps,
				!searchNotInDescription && !searchOnlyInName && !searchOnlyInTags && !searchOnlyInDeps,
				!searchNotInTags && !searchOnlyInName && !searchOnlyInDescription && !searchOnlyInDeps,
				!searchNotInDeps && !searchOnlyInName && !searchOnlyInDescription && !searchOnlyInTags,
			)
		}
		if len(results.Tools) == 0 {
			logging.Info.Printfln("No tools found for term <%s>", strings.Join(args, " "))
			return nil
		}

		switch searchOutputFormat {
		case "table":
			results.List(cmd.OutOrStdout())
		case "name":
			for _, tool := range results.Tools {
				fmt.Println(tool.Name)
			}
		case "json":
			data, err := json.Marshal(results)
			if err != nil {
				return fmt.Errorf("failed to marshal to json: %s", err)
			}
			fmt.Println(string(data))
		}

		return nil
	},
}
