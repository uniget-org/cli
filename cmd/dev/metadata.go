package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/pkg/git"
	"gitlab.com/uniget-org/cli/pkg/logging"
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
	metadataCmd.AddCommand(metadataChangesCmd)

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
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", data)

		} else {
			err := metadata.WriteMetadata(metadataFileName)
			if err != nil {
				return fmt.Errorf("failed to write metadata: %s", err)
			}
		}

		return nil
	},
}

var metadataChangesCmd = &cobra.Command{
	Use:     "changes",
	Aliases: []string{},
	Short:   "Collect metadata changes",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		metadata, err := tool.NewMetadataFromRegistry(registryHost, repositoryPrefix, "main")
		if err != nil {
			return fmt.Errorf("error loading metadata: %s", err)
		}
		logging.Info.Printfln("Metadata revision %s", metadata.Revision)

		forge := git.NewGitHubGitForge()
		//forge, err := git.NewGitLabGitForge()
		//if err != nil {
		//	return fmt.Errorf("unable to load gitlab client: %s", err)
		//}

		changes, err := forge.GetCommitChanges(metadata.Revision)
		if err != nil {
			return fmt.Errorf("error getting commit changes: %s", err)
		}
		for _, change := range changes.Changes {
			logging.Info.Printfln("change: %+v", change)
		}

		return nil
	},
}
