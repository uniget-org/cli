package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

var (
	metadataFileName = "metadata.json"
	metadataStdOut   = false
)

func initMetadataCmd() {
	metadataCreateCmd.Flags().StringVarP(&metadataFileName, "file", "f", metadataFileName, "Metadata file")
	metadataCreateCmd.Flags().BoolVarP(&metadataStdOut, "stdout", "o", metadataStdOut, "Output metadata to stdout")
	metadataCmd.AddCommand(metadataCreateCmd)

	rootCmd.AddCommand(metadataCmd)
}

var metadataCmd = &cobra.Command{
	Use:     "metadata",
	Aliases: []string{},
	Short:   "Work with metadata",
	Args:    cobra.NoArgs,
}

var metadataCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{},
	Short:   "Create metadata",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		metadata, err := tool.NewMetadataFromDirectory(unigetToolsDirectory + "/tools")
		if err != nil {
			return fmt.Errorf("error creating metadata: %s", err)
		}

		if metadataStdOut {
			data, err := json.Marshal(metadata)
			if err != nil {
				return nil
			}
			fmt.Fprintf(os.Stdout, "%s\n", data)

		} else {
			metadata.WriteMetadata(metadataFileName)
		}

		return nil
	},
}
