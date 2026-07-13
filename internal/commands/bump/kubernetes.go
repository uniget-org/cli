package bump

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/parse"
)

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

func processKubernetesFileCmd(cmd *cobra.Command, args []string) error {
	err := parse.BumpKubernetesFile(bumpKubernetesFileName, tools, outputCallback)
	if err != nil {
		return fmt.Errorf("failed to bump kubernetes file: %w", err)
	}

	return nil
}
