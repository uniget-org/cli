package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initImportCmd() {
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:     "import",
	Aliases: []string{},
	Short:   "Start managing existing binaries",
	Long:    header + "\nStart managing existing binaries",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		var err error

		pbar, _ := pterm.DefaultProgressbar.
			WithTotal(len(tools.Tools)).
			WithTitle("Checking").
			WithRemoveWhenDone().
			Start()

		importableTools := make([]huh.Option[string], 0)
		for _, tool := range tools.Tools {
			err = tool.UpdateStatus(viper.GetString("prefix"), viper.GetString("target"), cacheDirectory, arch, altArch)
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
