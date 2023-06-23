package main

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/nicholasdille/docker-setup/pkg/archive"
	"github.com/nicholasdille/docker-setup/pkg/containers"
	"github.com/nicholasdille/docker-setup/pkg/tool"

	"github.com/regclient/regclient/types/blob"
)

func initUpdateCmd() {
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

		if len(oldTools.Tools) > 0 {
			for _, tool := range tools.Tools {
				oldTool, _ := oldTools.GetByName(tool.Name)
				log.Tracef("Got tool for %s: %v\n", tool.Name, oldTool)

				if oldTool == nil {
					pterm.Info.Printfln("New %s v%s", tool.Name, tool.Version)

				} else if tool.Version != oldTool.Version {
					pterm.Info.Printfln("Update %s %s -> %s", tool.Name, oldTool.Version, tool.Version)
				}
			}
		}

		return nil
	},
}

func downloadMetadata() error {
	assertCacheDirectory()
	err := containers.GetManifest(registryImagePrefix+"metadata:main", altArch, func(blob blob.Reader) error {
		log.Tracef("Changing directory to %s", prefix+"/"+cacheDirectory)
		err := os.Chdir(prefix + "/" + cacheDirectory)
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", prefix+"/"+cacheDirectory, err)
		}

		log.Tracef("Extracting archive to %s", prefix+"/"+cacheDirectory)
		err = archive.ExtractTarGz(blob)
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
