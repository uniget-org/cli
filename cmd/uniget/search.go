package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

var (
	onlySearchInName        bool
	noSearchInName          bool
	onlySearchInDescription bool
	noSearchInDescription   bool
	onlySearchInTags        bool
	noSearchInTags          bool
	onlySearchInDeps        bool
	noSearchInDeps          bool
	output                  string
)

func initSearchCmd() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolVar(&onlySearchInName, "only-names", false, "Search only in names")
	searchCmd.Flags().BoolVar(&noSearchInName, "no-names", false, "Do not search in names")
	searchCmd.Flags().BoolVar(&onlySearchInDescription, "only-description", false, "Search only in description")
	searchCmd.Flags().BoolVar(&noSearchInDescription, "no-description", false, "Do not search in description")
	searchCmd.Flags().BoolVar(&onlySearchInTags, "only-tags", false, "Search only on tags")
	searchCmd.Flags().BoolVar(&noSearchInTags, "no-tags", false, "Do not search in tags")
	searchCmd.Flags().BoolVar(&onlySearchInDeps, "only-deps", false, "Search only in dependencies")
	searchCmd.Flags().BoolVar(&noSearchInDeps, "no-deps", false, "Do not search in dependencies")
	searchCmd.Flags().StringVar(&output, "output", "table", "Output format (table, name, json)")

	searchCmd.MarkFlagsMutuallyExclusive("only-names", "no-names")
	searchCmd.MarkFlagsMutuallyExclusive("only-description", "no-description")
	searchCmd.MarkFlagsMutuallyExclusive("only-tags", "no-tags")
	searchCmd.MarkFlagsMutuallyExclusive("only-deps", "no-deps")
}

var searchCmd = &cobra.Command{
	Use:     "search <term>",
	Aliases: []string{"s"},
	Short:   "Search for tools",
	Long:    header + "\nSearch for tools",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("update") {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		if output != "table" && output != "name" && output != "json" {
			return fmt.Errorf("error: output format %s not supported", output)
		}

		if (onlySearchInName && onlySearchInTags) ||
			(onlySearchInName && onlySearchInDeps) ||
			(onlySearchInName && noSearchInDescription) ||
			(onlySearchInDescription && onlySearchInTags) ||
			(onlySearchInDescription && onlySearchInDeps) ||
			(onlySearchInTags && onlySearchInDeps) {
			return fmt.Errorf("error: Can only process one of only-names, only-description, only-tags and only-deps at the same time")
		}

		results := tools.Find(
			args[0],
			!noSearchInName && !onlySearchInDescription && !onlySearchInTags && !onlySearchInDeps,
			!noSearchInDescription && !onlySearchInName && !onlySearchInTags && !onlySearchInDeps,
			!noSearchInTags && !onlySearchInName && !onlySearchInDescription && !onlySearchInDeps,
			!noSearchInDeps && !onlySearchInName && !onlySearchInDescription && !onlySearchInTags,
		)
		if len(results.Tools) == 0 {
			logging.Info.Printfln("No tools found for term %s", args[0])
			return nil
		}

		switch output {
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
