package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uniget-org/cli/pkg/logging"
)

func initSearchCmd() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().Bool("only-names", false, "Search only in names")
	searchCmd.Flags().Bool("no-names", false, "Do not search in names")
	searchCmd.Flags().Bool("only-description", false, "Search only in description")
	searchCmd.Flags().Bool("no-description", false, "Do not search in description")
	searchCmd.Flags().Bool("only-tags", false, "Search only on tags")
	searchCmd.Flags().Bool("no-tags", false, "Do not search in tags")
	searchCmd.Flags().Bool("only-deps", false, "Search only in dependencies")
	searchCmd.Flags().Bool("no-deps", false, "Do not search in dependencies")

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

		onlySearchInName, err := cmd.Flags().GetBool("only-names")
		if err != nil {
			return fmt.Errorf("error retrieving only-names flag: %s", err)
		}
		noSearchInName, err := cmd.Flags().GetBool("no-names")
		if err != nil {
			return fmt.Errorf("error retrieving no-names flag: %s", err)
		}
		onlySearchInDescription, err := cmd.Flags().GetBool("only-description")
		if err != nil {
			return fmt.Errorf("error retrieving only-description flag: %s", err)
		}
		noSearchInDescription, err := cmd.Flags().GetBool("no-description")
		if err != nil {
			return fmt.Errorf("error retrieving no-description flag: %s", err)
		}
		onlySearchInTags, err := cmd.Flags().GetBool("only-tags")
		if err != nil {
			return fmt.Errorf("error retrieving only-tags flag: %s", err)
		}
		noSearchInTags, err := cmd.Flags().GetBool("no-tags")
		if err != nil {
			return fmt.Errorf("error retrieving no-tags flag: %s", err)
		}
		onlySearchInDeps, err := cmd.Flags().GetBool("only-deps")
		if err != nil {
			return fmt.Errorf("error retrieving only-deps flag: %s", err)
		}
		noSearchInDeps, err := cmd.Flags().GetBool("no-deps")
		if err != nil {
			return fmt.Errorf("error retrieving no-deps flag: %s", err)
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

		results.List(cmd.OutOrStdout())

		return nil
	},
}
