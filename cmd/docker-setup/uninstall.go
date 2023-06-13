package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if fileExists(prefix + "/" + metadataFile) {
			log.Tracef("Loaded metadata file from %s", prefix+"/"+metadataFile)
			loadMetadata()
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		assertWritableTarget()
		assertLibDirectory()

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("unable to find tool %s: %s", args[0], err)
		}
		if fileExists(prefix + "/" + libDirectory + "/manifests/" + tool.Name + ".txt") {
			// TODO: Remove all files listes in /var/lib/docker-setup/manifests/<tool>.txt
			tool.RemoveMarkerFile(prefix + "/" + cacheDirectory)
			// TODO: Remove prefix + "/" + libDirectory + "/manifests/" + tool.Name + ".txt"
			// TODO: Remove prefix + "/" + libDirectory + "/manifests/" + tool.Name + ".json"
		} else {
			return fmt.Errorf("tool %s does not have a manifest file. Is it installed?", tool.Name)
		}

		return nil
	},
}
