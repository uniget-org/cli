package main

import (
	"context"
	"io"

	"github.com/google/safearchive/tar"
	"github.com/spf13/cobra"

	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

func initMetadataCmd() {
	metadataCmd.AddCommand(downloadMetadataCmd)
	rootCmd.AddCommand(metadataCmd)
}

var metadataCmd = &cobra.Command{
	Use: "metadata",
	Aliases: []string{
		"meta",
	},
	Short:  "Manage metadata",
	Long:   constants.Header + "\nManage metadata",
	Hidden: true,
	Args:   cobra.NoArgs,
}

var downloadMetadataCmd = &cobra.Command{
	Use: "download",
	Aliases: []string{
		"down",
		"d",
		"get",
		"fetch",
	},
	Short: "Download metadata",
	Long:  constants.Header + "\nDownload metadata",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		configuration.AssertCacheDirectory()
		t, err := containers.FindToolRef([]string{constants.Registry}, []string{constants.ImageRepository}, "metadata", constants.MetadataImageTag)
		if err != nil {
			return handleError(err, "Error finding metadata: %s", err)
		}
		rc := containers.GetRegclient()
		defer func() {
			err := rc.Close(context.Background(), t.GetRef())
			if err != nil {
				logging.Warning.Printfln("Failed to close registry client: %s", err)
			}
		}()

		progressReader := tui.NewProgressReader(nil, nil)
		err = containers.GetFirstLayerFromRegistry(context.Background(), rc, t.GetRef(), progressReader, func(reader io.ReadCloser) error {
			return archive.ProcessTarContents(reader, func(reader *tar.Reader, header *tar.Header) error {
				if header.Typeflag == tar.TypeReg && header.Name == "metadata.json" {
					_, err = io.Copy(cmd.OutOrStdout(), io.NopCloser(reader))
					if err != nil {
						return handleError(err, "Error writing metadata to stdout")
					}
				}

				return nil
			})

		})
		if err != nil {
			return handleError(err, "Error getting first layer from registry")
		}

		return nil
	},
}
