package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/config"
	"gitlab.com/uniget-org/cli/internal/constants"
)

var debugCmd *cobra.Command

func NewDebugCommand(configurationProvider func() *config.Config) *cobra.Command {
	debugCmd = &cobra.Command{
		Use:     "debug",
		Aliases: []string{},
		Short:   "Debug parameters",
		Long:    constants.Header + "\nDebug parameters",
		Hidden:  true,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration := configurationProvider()

			//nolint:errcheck
			fmt.Fprintf(cmd.OutOrStdout(), "configuration: %s\n", configuration)

			return nil
		},
	}

	return debugCmd
}
