package main

import (
	"fmt"
	"os"

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uniget-org/cli/pkg/logging"
)

var manDirectory string

func initManpagesCmd() {
	manpagesCmd.Flags().StringVar(&manDirectory, "path", "share/man", "Path to store manpages in (relative paths resolves using target directory)")

	rootCmd.AddCommand(manpagesCmd)
}

var manpagesCmd = &cobra.Command{
	Use:     "manpages",
	Aliases: []string{"man", "manpage"},
	Short:   "Generate manpages",
	Long:    header + "\nGenerate manpages",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if (manDirectory[0:1] != "/") && (manDirectory[0:1] != ".") {
			manDirectory = fmt.Sprintf("%s/%s", "/"+viper.GetString("target"), manDirectory)
		}
		logging.Debugf("Using base directory for manpages: %s", manDirectory)

		err := writeManpage(rootCmd, "", manDirectory)
		if err != nil {
			return fmt.Errorf("failed to create manpage: %w", err)
		}

		for _, cobraCmd := range rootCmd.Commands() {
			logging.Info.Printfln("Generating manpage for %s...", cobraCmd.Name())

			err := writeManpage(cobraCmd, cobraCmd.Name(), manDirectory)
			if err != nil {
				return fmt.Errorf("failed to create manpage: %w", err)
			}
		}

		return nil
	},
}

func writeManpage(cobraCmd *cobra.Command, name string, manDirectory string) error {
	manPage, err := mcobra.NewManPage(1, cobraCmd)
	if err != nil {
		return fmt.Errorf("unable to create manpage: %w", err)
	}

	manPage = manPage.WithSection(
		"Copyright", "(C) 2023 Nicholas Dille.\n"+
			"Released under MIT license.",
	)

	var fileName string
	dirName := fmt.Sprintf("%s/man1", manDirectory)
	if !directoryExists(dirName) {
		err := os.MkdirAll(dirName, 0755) // #nosec G301 -- Directory need to be accessible by all users
		if err != nil {
			return fmt.Errorf("failed to create manpage directory: %w", err)
		}
	}
	if name == "" {
		fileName = fmt.Sprintf("%s/%s.1", dirName, projectName)
	} else {
		fileName = fmt.Sprintf("%s/%s-%s.1", dirName, projectName, name)
	}

	file, err := os.Create(fileName) // #nosec G304 -- This is exactly the value proposition of this command
	if err != nil {
		return fmt.Errorf("failed to create manpage: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logging.Warning.Printfln("failed to close manpage file: %s", err)
		}
	}()

	_, err = file.WriteString(manPage.Build(roff.NewDocument()))
	return err
}
