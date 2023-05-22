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
		assertCacheDirectory()
		containers.GetManifest(registryImagePrefix + "metadata:main", alt_arch, func (blob blob.Reader) error {
			err := os.Chdir(cacheDirectory)
			if err != nil {
				fmt.Printf("Error changing directory to %s: %s\n", cacheDirectory, err)
				os.Exit(1)
			}
			archive.ExtractTarGz(blob)
			return nil
		})

		loadMetadata()
	},
}
