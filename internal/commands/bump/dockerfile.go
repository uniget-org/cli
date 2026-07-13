package bump

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/parse"
)

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

func processBumpDockerfileCmd(cmd *cobra.Command, args []string) error {
	err := parse.BumpDockerfile(bumpDockerfileName, tools, outputCallback)
	if err != nil {
		return fmt.Errorf("failed to bump dockerfile: %w", err)
	}

	return nil
}
