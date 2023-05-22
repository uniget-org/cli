package main

import (
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
	RunE:      func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		assertWritableTarget()
		assertLibDirectory()
		// Remove all files listes in /var/lib/docker-setup/manifests/<tool>.txt
		// tool.RemoveMarkerFile()

		return nil
	},
}
