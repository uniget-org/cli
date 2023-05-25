package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func initUninstallCmd() {
	rootCmd.AddCommand(uninstallCmd)
}

var uninstallCmd = &cobra.Command{
	Use:       "uninstall",
	Aliases:   []string{"u"},
	Short:     "Uninstall tool",
	Long:      header + "\nUninstall tools",
	Args:      cobra.ExactArgs(1),
	ValidArgs: tools.GetNames(),
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		assertWritableTarget()
		assertLibDirectory()

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("unable to find tool %s: %s", args[0], err)
		}
		if fileExists(libDirectory + "/manifests/" + tool.Name + ".txt") {
			// Remove all files listes in /var/lib/docker-setup/manifests/<tool>.txt
			tool.RemoveMarkerFile(cacheDirectory)
			// Remove libDirectory + "/manifests/" + tool.Name + ".txt"
			// Remove libDirectory + "/manifests/" + tool.Name + ".json"
		} else {
			return fmt.Errorf("tool %s does not have a manifest file. Is it installed?", tool.Name)
		}

		return nil
	},
}
