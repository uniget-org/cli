package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/google/safearchive/tar"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/logging"
	"github.com/uniget-org/cli/pkg/tool"
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

		newUnigetVersion := ""
		if !quiet && len(oldTools.Tools) > 0 {
			for _, tool := range tools.Tools {
				oldTool, _ := oldTools.GetByName(tool.Name)

				if tool.Name == "uniget" && tool.Version != version {
					newUnigetVersion = tool.Version
				}

				if oldTool == nil {
					logging.Info.Printfln("New %s %s", tool.Name, tool.Version)

				} else if tool.Version != oldTool.Version {
					logging.Info.Printfln("Update %s to %s", tool.Name, tool.Version)
				}
			}
		}

		if len(newUnigetVersion) > 0 {
			prefix := pterm.NewStyle(pterm.FgBlack, pterm.BgYellow)
			suffix := pterm.NewStyle(pterm.FgWhite)
			prefix.Println()
			prefix.Print(" NEWS ")
			suffix.Printfln(" Update to uniget %s by running 'uniget self-upgrade'", newUnigetVersion)
		}

		return nil
	},
}

func downloadMetadata() error {
	assertCacheDirectory()
	t, err := containers.FindToolRef([]string{registry}, []string{imageRepository}, "metadata", metadataImageTag)
	if err != nil {
		return fmt.Errorf("error finding metadata: %s", err)
	}
	rc := containers.GetRegclient()
	defer func() {
		err := rc.Close(context.Background(), t.GetRef())
		if err != nil {
			logging.Warning.Printfln("error closing registry client: %s", err)
		}
	}()

	logging.Debugf("Changing directory to %s", viper.GetString("prefix")+"/"+cacheDirectory)
	err = os.Chdir(viper.GetString("prefix") + "/" + cacheDirectory)
	if err != nil {
		return fmt.Errorf("error changing directory to %s: %s", viper.GetString("prefix")+"/"+cacheDirectory, err)
	}

	logging.Debugf("Extracting archive to %s", viper.GetString("prefix")+"/"+cacheDirectory)
	err = containers.GetFirstLayerFromRegistry(context.Background(), rc, t.GetRef(), func(reader io.ReadCloser) error {
		err := archive.ProcessTarContents(reader, func(reader *tar.Reader, header *tar.Header) error {
			err := archive.CallbackExtractTarItem(reader, header)
			if err != nil {
				return fmt.Errorf("error extracting tar item: %s", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error processing tar contents: %s", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error getting first layer from registry: %s", err)
	}

	return nil
}

func loadMetadata() error {
	var err error
	tools, err = tool.LoadFromFile(viper.GetString("prefix") + "/" + metadataFile)
	if err != nil {
		return fmt.Errorf("failed to load metadata from file %s: %s", viper.GetString("prefix")+"/"+metadataFile, err)
	}

	return nil
}
