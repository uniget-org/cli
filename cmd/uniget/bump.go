package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/logging"
	"github.com/uniget-org/cli/pkg/parse"
)

var (
	bumpDockerfileName     = "Dockerfile"
	bumpComposeFileName    = "compose.yaml"
	bumpKubernetesFileName = ""
)

func initBumpCmd() {
	bumpDockerfileCmd.Flags().StringVarP(&bumpDockerfileName, "file", "f", bumpDockerfileName, "Path to Dockerfile")
	bumpComposeCmd.Flags().StringVarP(&bumpComposeFileName, "file", "f", bumpComposeFileName, "Path to compose file")
	bumpKubernetesCmd.Flags().StringVarP(&bumpKubernetesFileName, "file", "f", bumpKubernetesFileName, "Path to Kubernetes manifest")

	err := bumpKubernetesCmd.MarkFlagRequired("file")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}

	bumpCmd.AddCommand(bumpDockerfileCmd)
	bumpCmd.AddCommand(bumpComposeCmd)
	bumpCmd.AddCommand(bumpKubernetesCmd)
	rootCmd.AddCommand(bumpCmd)
}

var bumpCmd = &cobra.Command{
	Use:   "bump",
	Short: "Bump tool versions",
	Long:  header + "\nBump tool versions",
	Args:  cobra.NoArgs,
}

var bumpDockerfileCmd = &cobra.Command{
	Use:   "dockerfile",
	Short: "Bump image references in a Dockerfile",
	Long:  header + "\nBump image references in a Dockerfile",
	Args:  cobra.NoArgs,
	RunE:  processBumpDockerfileCmd,
}

var bumpComposeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Bump image references in a compose file",
	Long:  header + "\nBump image references in a compose file",
	Args:  cobra.NoArgs,
	RunE:  processComposeFileCmd,
}

var bumpKubernetesCmd = &cobra.Command{
	Use:     "kubernetes",
	Aliases: []string{"k8s"},
	Short:   "Bump image references in a Kubernetes manifest",
	Long:    header + "\nBump image references in a Kubernetes manifest",
	Args:    cobra.NoArgs,
	RunE:    processKubernetesFileCmd,
}

func processBumpDockerfileCmd(cmd *cobra.Command, args []string) error {
	assertMetadataFileExists()
	assertMetadataIsLoaded()

	err := parse.BumpDockerfile(bumpDockerfileName, &tools)
	if err != nil {
		return fmt.Errorf("failed to bump dockerfile: %w", err)
	}

	return nil
}

func processComposeFileCmd(cmd *cobra.Command, args []string) error {
	assertMetadataFileExists()
	assertMetadataIsLoaded()

	err := parse.BumpComposeFile(bumpComposeFileName, &tools)
	if err != nil {
		return fmt.Errorf("failed to bump compose file: %w", err)
	}

	return nil
}

func processKubernetesFileCmd(cmd *cobra.Command, args []string) error {
	assertMetadataFileExists()
	assertMetadataIsLoaded()

	err := parse.BumpKubernetesFile(bumpKubernetesFileName, &tools)
	if err != nil {
		return fmt.Errorf("failed to bump kubernetes file: %w", err)
	}

	return nil
}
