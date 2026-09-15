package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/security"
)

var ()

func initVerifyCmd() {
	verifyCmd.AddCommand(verifyMetadataCmd)
	verifyCmd.AddCommand(verifyToolCmd)
	rootCmd.AddCommand(verifyCmd)
}

var verifyCmd = &cobra.Command{
	Use: "verify",
	Aliases: []string{
		"v",
	},
	Short:   "Verify signatures",
	Long:    constants.Header + "\nVerify signatures",
	GroupID: "helper",
	Hidden:  true,
}

var verifyMetadataCmd = &cobra.Command{
	Use: "metadata",
	Aliases: []string{
		"m",
	},
	Short: "Verify metadata signatures",
	Long:  constants.Header + "\nVerify metadata signatures",
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := configuration.GetMetadataFile()

		_, err := security.VerifySigstoreBundleForArtifact(
			filename,
			filename+".sigstore.json",
			constants.SigstoreIssuer,
			"",
			"",
			"https://github\\.com/uniget-org/tools/\\.github/workflows/[^.]+\\.yml@refs/heads/main",
		)
		if err != nil {
			return fmt.Errorf("error verifying sigstore bundle for metadata: %w", err)
		}

		logging.Info.Printfln("Metadata signature verified successfully")

		return nil
	},
}

var verifyToolCmd = &cobra.Command{
	Use: "tool",
	Aliases: []string{
		"t",
	},
	Short: "Verify tool signatures",
	Long:  constants.Header + "\nVerify tool signatures",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		tools, err = configuration.EnsureMetadata()
		if err != nil {
			return fmt.Errorf("error ensuring metadata: %s", err)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("error getting tool by name: %w", err)
		}
		registries, repositories := tool.GetSourcesWithFallback(constants.Registry, constants.ImageRepository)
		ref, err := containers.FindToolRef(registries, repositories, tool.Name, "latest")
		if err != nil {
			return fmt.Errorf("error finding tool reference: %w", err)
		}
		logging.Debugf("Built reference %s", ref)

		err = containers.VerifyContainerImageSignature(ref, constants.SigstoreIssuer, "", "", constants.SigstoreSubjectRegexp)
		if err != nil {
			return fmt.Errorf("error verifying tool signature: %w", err)
		}
		logging.Success.Printfln("Signature verified successfully for %s", tool.Name)

		return nil
	},
}
