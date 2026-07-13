package bump

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/parse"
)

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

func processGitlabCiFileCmd(cmd *cobra.Command, args []string) error {
	err := parse.BumpGitlabCiFile(bumpGitLabCiFileName, tools, outputCallback)
	if err != nil {
		return fmt.Errorf("failed to bump GitLab CI file: %w", err)
	}

	return nil
}
