package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func initSearchCmd() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolP("only-names", "", false, "Search only in names")
	searchCmd.Flags().BoolP("no-names",   "", false, "Do not search in names")
	searchCmd.Flags().BoolP("only-tags",  "", false, "Search only on tags")
	searchCmd.Flags().BoolP("no-tags",    "", false, "Do not search in tags")
	searchCmd.Flags().BoolP("only-deps",  "", false, "Search only in dependencies")
	searchCmd.Flags().BoolP("no-deps",    "", false, "Do not search in dependencies")
}

var searchCmd = &cobra.Command{
	Use:     "search <term>",
	Aliases: []string{"s"},
	Short:   "Search for tools",
	Long:    header + "\nSearch for tools",
	Args:    cobra.ExactArgs(1),
	RunE:    func(cmd *cobra.Command, args []string) error {
		onlySearchInName, err := cmd.Flags().GetBool("only-names")
		if err != nil {
			return fmt.Errorf("Error retrieving only-names flag: %s\n", err)
		}
		noSearchInName, err := cmd.Flags().GetBool("no-names")
		if err != nil {
			return fmt.Errorf("Error retrieving no-names flag: %s\n", err)
		}
		onlySearchInTags, err := cmd.Flags().GetBool("only-tags")
		if err != nil {
			return fmt.Errorf("Error retrieving only-tags flag: %s\n", err)
		}
		noSearchInTags, err := cmd.Flags().GetBool("no-tags")
		if err != nil {
			return fmt.Errorf("Error retrieving no-tags flag: %s\n", err)
		}
		onlySearchInDeps, err := cmd.Flags().GetBool("only-deps")
		if err != nil {
			return fmt.Errorf("Error retrieving only-deps flag: %s\n", err)
		}
		noSearchInDeps, err := cmd.Flags().GetBool("no-deps")
		if err != nil {
			return fmt.Errorf("Error retrieving no-deps flag: %s\n", err)
		}

		if onlySearchInName && noSearchInName {
			return fmt.Errorf("Error: Cannot process only-names and no-names at the same time\n")
		}
		if onlySearchInTags && noSearchInTags {
			return fmt.Errorf("Error: Cannot process only-tags and no-tags at the same time\n")
		}
		if onlySearchInDeps && noSearchInDeps {
			return fmt.Errorf("Error: Cannot process only-deps and no-deps at the same time\n")
		}

		if (onlySearchInName && onlySearchInTags) ||
			(onlySearchInName && onlySearchInDeps) ||
			(onlySearchInTags && onlySearchInDeps) {
			return fmt.Errorf("Error: Can only process one of only-names, only-tags and only-deps at the same time\n")
		}

		results := tools.Find(
			args[0],
			!noSearchInName && !onlySearchInTags && !onlySearchInDeps,
			!noSearchInTags && !onlySearchInName && !onlySearchInDeps,
			!noSearchInDeps && !onlySearchInName && !onlySearchInTags,
		)
		if len(results.Tools) == 0 {
			fmt.Printf("No tools found for term %s\n", args[0])

		} else {
			results.List()
		}

		return nil
	},
}
