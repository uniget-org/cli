package bump

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/parse"
)

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

func processComposeFileCmd(cmd *cobra.Command, args []string) error {
	err := parse.BumpComposeFile(bumpComposeFileName, tools, outputCallback)
	if err != nil {
		return fmt.Errorf("failed to bump compose file: %w", err)
	}

	return nil
}
