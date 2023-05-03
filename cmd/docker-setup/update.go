package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nicholasdille/docker-setup/pkg/archive"
	"github.com/nicholasdille/docker-setup/pkg/containers"

	"github.com/regclient/regclient/types/blob"
)

func initUpdateCmd() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update tool manifest",
	Long:    header + "\nUpdate tool manifest",
	Args:    cobra.NoArgs,
	Run:     func(cmd *cobra.Command, args []string) {
		// TODO: cacheDiectory is writable

		containers.GetManifest("ghcr.io/nicholasdille/docker-setup/metadata:main", alt_arch, func (blob blob.Reader) error {
			os.Chdir(cacheDirectory)
			archive.ExtractTarGz(blob)
			return nil
		})
	},
}
