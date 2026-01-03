package main

import (
	"github.com/spf13/cobra"
)

func initManifestCmd() {
	rootCmd.AddCommand(manifestCmd)
}

var manifestCmd = &cobra.Command{
	Use:     "manifest",
	Aliases: []string{},
	Short:   "Work with manifests",
	Args:    cobra.NoArgs,
}
