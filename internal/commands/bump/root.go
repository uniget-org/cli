package bump

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/config"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

var (
	//configuration          *config.Config
	tools                  *tool.Tools
	bumpCmd                *cobra.Command
	bumpDockerfileName     = "Dockerfile"
	bumpComposeFileName    = "compose.yaml"
	bumpKubernetesFileName = ""
	bumpGitLabCiFileName   = ".gitlab-ci.yml"
)

var outputCallback = func(toolName string, oldVersion string, newVersion string) {
	pterm.NewStyle(pterm.FgBlack, pterm.BgYellow).Print("  BUMP  ")
	pterm.NewStyle(pterm.FgWhite).Printfln(" %s %s", toolName, newVersion)
}

func NewBumpCommand(configurationProvider func() *config.Config, toolsProvider func() *tool.Tools) *cobra.Command {
	//configuration = configurationProvider()
	tools = toolsProvider()

	bumpCmd = &cobra.Command{
		Use: "bump",
		Aliases: []string{
			"b",
		},
		Short: "Bump tool versions",
		Long:  constants.Header + "\nBump tool versions",
		Args:  cobra.NoArgs,
	}

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

	return bumpCmd
}
