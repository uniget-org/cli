package main

import (
	"fmt"
	"os"

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var manDirectory string

func initManCmd() {
	manCmd.Flags().StringVar(&manDirectory, "path", ".", "Path to store manpages in")

	rootCmd.AddCommand(manCmd)
}

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate manpages",
	Long:  header + "\nGenerate manpages",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := writeManpage(rootCmd, "", manDirectory)
		if err != nil {
			return fmt.Errorf("failed to create manpage: %w", err)
		}

		for _, cobraCmd := range rootCmd.Commands() {
			pterm.Info.Printfln("Generating manpage for %s...", cobraCmd.Name())

			err := writeManpage(cobraCmd, cobraCmd.Name(), manDirectory)
			if err != nil {
				return fmt.Errorf("failed to create manpage: %w", err)
			}
		}

		return nil
	},
}

func writeManpage(cobraCmd *cobra.Command, name string, path string) error {
	manPage, err := mcobra.NewManPage(1, cobraCmd)
	if err != nil {
		panic(err)
	}

	manPage = manPage.WithSection(
		"Copyright", "(C) 2023 Nicholas Dille.\n"+
			"Released under MIT license.",
	)

	var fileName string
	if name == "" {
		fileName = fmt.Sprintf("%s/%s.1", manDirectory, projectName)
	} else {
		fileName = fmt.Sprintf("%s/%s-%s.1", manDirectory, projectName, name)
	}

	file, err := os.Create(fileName) // #nosec G304 -- This is exactly the value proposition of this command
	if err != nil {
		return fmt.Errorf("failed to create manpage: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(manPage.Build(roff.NewDocument()))
	return err
}
