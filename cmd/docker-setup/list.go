package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCmd() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "get"},
	Short:   "List tools",
	Long:    header + "\nList tools",
	Args:    cobra.NoArgs,
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

		tools.List()

		return nil
	},
}
