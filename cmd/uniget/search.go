package main

import (
	"fmt"
	"os"

	"atomicgo.dev/keyboard/keys"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/logging"
	"golang.org/x/term"
)

func initSearchCmd() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().Bool("only-names", false, "Search only in names")
	searchCmd.Flags().Bool("no-names", false, "Do not search in names")
	searchCmd.Flags().Bool("only-tags", false, "Search only on tags")
	searchCmd.Flags().Bool("no-tags", false, "Do not search in tags")
	searchCmd.Flags().Bool("only-deps", false, "Search only in dependencies")
	searchCmd.Flags().Bool("no-deps", false, "Do not search in dependencies")
}

var searchCmd = &cobra.Command{
	Use:     "search <term>",
	Aliases: []string{"s"},
	Short:   "Search for tools",
	Long:    header + "\nSearch for tools",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if onlySearchInName && noSearchInName {
			return fmt.Errorf("error: Cannot process only-names and no-names at the same time")
		}
		if onlySearchInTags && noSearchInTags {
			return fmt.Errorf("error: Cannot process only-tags and no-tags at the same time")
		}
		if onlySearchInDeps && noSearchInDeps {
			return fmt.Errorf("error: Cannot process only-deps and no-deps at the same time")
		}

		if (onlySearchInName && onlySearchInTags) ||
			(onlySearchInName && onlySearchInDeps) ||
			(onlySearchInTags && onlySearchInDeps) {
			return fmt.Errorf("error: Can only process one of only-names, only-tags and only-deps at the same time")
		}

		results := tools.Find(
			args[0],
			!noSearchInName && !onlySearchInTags && !onlySearchInDeps,
			!noSearchInTags && !onlySearchInName && !onlySearchInDeps,
			!noSearchInDeps && !onlySearchInName && !onlySearchInTags,
		)
		if len(results.Tools) == 0 {
			logging.Info.Printfln("No tools found for term %s", args[0])
			return nil
		}

		results.List()

		if noInteractive || !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
			return nil
		}

		fmt.Println()
		var options []string
		for _, tool := range results.Tools {
			options = append(options, tool.Name)
		}
		printer := pterm.DefaultInteractiveMultiselect.WithOptions(options)
		printer.DefaultText = "Select tools to install"
		printer.Filter = false
		printer.KeyConfirm = keys.Enter
		printer.KeySelect = keys.Space
		printer.Checkmark = &pterm.Checkmark{Checked: "✓", Unchecked: " "}
		selectedOptions, _ := printer.Show()
		if len(selectedOptions) > 0 {
			err := installToolsByName(selectedOptions, false, false, false, false, false)
			if err != nil {
				return err
			}
		}

		return nil
	},
}
