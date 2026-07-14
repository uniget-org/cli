package main

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
)

func initImportCmd() {
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:     "import",
	Aliases: []string{},
	Short:   "Start managing existing binaries",
	Long:    constants.Header + "\nStart managing existing binaries",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if configuration.AutoUpdate {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		configuration.AssertMetadataFileExists()
		assertMetadataIsLoaded()

		var err error

		pbar, _ := pterm.DefaultProgressbar.
			WithTotal(len(tools.Tools)).
			WithTitle("Checking").
			WithRemoveWhenDone().
			Start()

		importableTools := make([]huh.Option[string], 0)
		for _, tool := range tools.Tools {
			err = tool.UpdateStatus(
				configuration.Prefix,
				configuration.Target,
				configuration.GetCacheDirectory(),
				configuration.Arch,
				configuration.AltArch,
			)
			if err != nil {
				return fmt.Errorf("failed to update status for tool %s: %s", tool.Name, err)
			}

			if tool.IsImportable() {
				importableTools = append(importableTools, huh.NewOption(tool.Name, tool.Name))
			}

			pbar.Increment()
		}

		toolsToImport := make([]string, 0)
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Tools to import").
					Description("Selected tools will be installed").
					Options(importableTools...).
					Height(10).
					Value(&toolsToImport),
			),
		)
		err = form.Run()
		if err != nil {
			return fmt.Errorf("failed to run form: %s", err)
		}

		plannedTools := tools.GetByNames(toolsToImport)
		err = installTools(cmd.OutOrStdout(), plannedTools, false, false, true, true, true)
		if err != nil {
			return fmt.Errorf("failed to import tools: %s", err)
		}

		return nil
	},
}
