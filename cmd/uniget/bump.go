package main

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/parse"
)

var (
	bumpDockerfileName     = "Dockerfile"
	bumpComposeFileName    = "compose.yaml"
	bumpKubernetesFileName = ""
	bumpGitLabCiFileName   = ".gitlab-ci.yml"
)

var outputCallback = func(toolName string, oldVersion string, newVersion string) {
	pterm.NewStyle(pterm.FgBlack, pterm.BgYellow).Print("  BUMP  ")
	pterm.NewStyle(pterm.FgWhite).Printfln(" %s %s", toolName, newVersion)
}

func initBumpCmd() {
	bumpDockerfileCmd.Flags().StringVarP(&bumpDockerfileName, "file", "f", bumpDockerfileName, "Path to Dockerfile")
	bumpComposeCmd.Flags().StringVarP(&bumpComposeFileName, "file", "f", bumpComposeFileName, "Path to compose file")
	bumpKubernetesCmd.Flags().StringVarP(&bumpKubernetesFileName, "file", "f", bumpKubernetesFileName, "Path to Kubernetes manifest")
	bumpGitlabCiCmd.Flags().StringVarP(&bumpGitLabCiFileName, "file", "f", bumpGitLabCiFileName, "Path to GitLab CI file")

	err := bumpKubernetesCmd.MarkFlagRequired("file")
	if err != nil {
		logging.Error.Printfln("Failed to mark flag as required: %v", err)
	}

	bumpCmd.AddCommand(bumpDockerfileCmd)
	bumpCmd.AddCommand(bumpComposeCmd)
	bumpCmd.AddCommand(bumpKubernetesCmd)
	bumpCmd.AddCommand(bumpGitlabCiCmd)
	rootCmd.AddCommand(bumpCmd)
}

var bumpCmd = &cobra.Command{
	Use: "bump",
	Aliases: []string{
		"b",
	},
	Short: "Bump tool versions",
	Long:  constants.Header + "\nBump tool versions",
	Args:  cobra.NoArgs,
}

var bumpDockerfileCmd = &cobra.Command{
	Use: "dockerfile",
	Aliases: []string{
		"docker",
		"df",
	},
	Short: "Bump image references in a Dockerfile",
	Long:  constants.Header + "\nBump image references in a Dockerfile",
	Args:  cobra.NoArgs,
	RunE:  processBumpDockerfileCmd,
}

var bumpComposeCmd = &cobra.Command{
	Use: "compose",
	Aliases: []string{
		"c",
		"docker-compose",
		"dc",
	},
	Short: "Bump image references in a compose file",
	Long:  constants.Header + "\nBump image references in a compose file",
	Args:  cobra.NoArgs,
	RunE:  processComposeFileCmd,
}

var bumpKubernetesCmd = &cobra.Command{
	Use: "kubernetes",
	Aliases: []string{
		"k",
		"k8s",
	},
	Short: "Bump image references in a Kubernetes manifest",
	Long:  constants.Header + "\nBump image references in a Kubernetes manifest",
	Args:  cobra.NoArgs,
	RunE:  processKubernetesFileCmd,
}

var bumpGitlabCiCmd = &cobra.Command{
	Use: "gitlab-ci",
	Aliases: []string{
		"gitlab",
		"gl",
	},
	Short: "Bump image references in a GitLab CI file",
	Long:  constants.Header + "\nBump image references in a GitLab CI file",
	Args:  cobra.NoArgs,
	RunE:  processGitlabCiFileCmd,
}

func processBumpDockerfileCmd(cmd *cobra.Command, args []string) error {
	assertMetadataFileExists()
	assertMetadataIsLoaded()

	err := parse.BumpDockerfile(bumpDockerfileName, &tools, outputCallback)
	if err != nil {
		return fmt.Errorf("failed to bump dockerfile: %w", err)
	}

	return nil
}

func processComposeFileCmd(cmd *cobra.Command, args []string) error {
	assertMetadataFileExists()
	assertMetadataIsLoaded()

	err := parse.BumpComposeFile(bumpComposeFileName, &tools, outputCallback)
	if err != nil {
		return fmt.Errorf("failed to bump compose file: %w", err)
	}

	return nil
}

func processKubernetesFileCmd(cmd *cobra.Command, args []string) error {
	assertMetadataFileExists()
	assertMetadataIsLoaded()

	err := parse.BumpKubernetesFile(bumpKubernetesFileName, &tools, outputCallback)
	if err != nil {
		return fmt.Errorf("failed to bump kubernetes file: %w", err)
	}

	return nil
}

func processGitlabCiFileCmd(cmd *cobra.Command, args []string) error {
	assertMetadataFileExists()
	assertMetadataIsLoaded()

	err := parse.BumpGitlabCiFile(bumpGitLabCiFileName, &tools, outputCallback)
	if err != nil {
		return fmt.Errorf("failed to bump GitLab CI file: %w", err)
	}

	return nil
}
