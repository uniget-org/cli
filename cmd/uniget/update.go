package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/logging"
	"github.com/uniget-org/cli/pkg/tool"

	"github.com/regclient/regclient/types/blob"
)

var quiet bool

func initUpdateCmd() {
	updateCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Do not print new tools")

	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update tool manifest",
	Long:  header + "\nUpdate tool manifest",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := downloadMetadata()
		if err != nil {
			return fmt.Errorf("error downloading metadata: %s", err)
		}

		oldTools := tools
		err = loadMetadata()
		if err != nil {
			return fmt.Errorf("error loading metadata: %s", err)
		}

		if !quiet && len(oldTools.Tools) > 0 {
			for _, tool := range tools.Tools {
				oldTool, _ := oldTools.GetByName(tool.Name)

				if oldTool == nil {
					logging.Info.Printfln("New %s v%s", tool.Name, tool.Version)

				} else if tool.Version != oldTool.Version {
					logging.Info.Printfln("Update %s %s -> %s", tool.Name, oldTool.Version, tool.Version)
				}
			}
		}

		return nil
	},
}

func downloadMetadata() error {
	assertCacheDirectory()
	err := containers.GetManifest(registryImagePrefix+"metadata:main", altArch, func(blob blob.Reader) error {
		logging.Debug.Printfln("Changing directory to %s", prefix+"/"+cacheDirectory)
		err := os.Chdir(prefix + "/" + cacheDirectory)
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", prefix+"/"+cacheDirectory, err)
		}

		logging.Debug.Printfln("Extracting archive to %s", prefix+"/"+cacheDirectory)
		err = archive.ExtractTarGz(blob, func(path string) string { return path })
		if err != nil {
			return fmt.Errorf("error extracting archive: %s", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error getting manifest: %s", err)
	}

	return nil
}

func loadMetadata() error {
	var err error
	tools, err = tool.LoadFromFile(prefix + "/" + metadataFile)
	if err != nil {
		return fmt.Errorf("failed to load metadata from file %s: %s", prefix+"/"+metadataFile, err)
	}

	return nil
}
