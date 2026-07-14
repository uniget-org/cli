package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/google/safearchive/tar"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/security"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

var updateQuiet bool
var updateShowAllTools bool

func initUpdateCmd() {
	updateCmd.Flags().BoolVarP(&updateQuiet, "quiet", "q", false, "Do not print new tools")
	updateCmd.Flags().BoolVar(&updateShowAllTools, "all", false, "Show all updates including tools that are not installed")

	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{},
	Short:   "Update tool manifest",
	Long:    constants.Header + "\nUpdate tool manifest",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		oldTools := tools

		err = downloadMetadata()
		if err != nil {
			return fmt.Errorf("error downloading metadata: %s", err)
		}

		err = loadMetadata()
		if err != nil {
			return fmt.Errorf("error loading metadata: %s", err)
		}

		var updatedTools tool.Tools
		var updatedInstalledTools tool.Tools
		var newTools tool.Tools
		newUnigetVersion := ""
		if len(oldTools.Tools) > 0 {
			for _, tool := range tools.Tools {
				oldTool, _ := oldTools.GetByName(tool.Name)

				if tool.Name == "uniget" && tool.Version > version {
					newUnigetVersion = tool.Version
				}

				if oldTool == nil {
					newTools.Tools = append(newTools.Tools, tool)

				} else if tool.Version != oldTool.Version {
					err := tool.UpdateStatus(
						configuration.Prefix,
						configuration.Target,
						configuration.GetCacheDirectory(),
						configuration.Arch,
						configuration.AltArch,
					)
					if err != nil {
						logging.Warning.Printfln("Error updating status for %s: %s", tool.Name, err)
					}

					updatedTools.Tools = append(updatedTools.Tools, tool)
					if tool.IsInstalled() {
						updatedInstalledTools.Tools = append(updatedInstalledTools.Tools, tool)
					}
				}
			}
		}

		if !updateQuiet {
			prefix := pterm.NewStyle(pterm.FgBlack, pterm.BgGreen)
			suffix := pterm.NewStyle(pterm.FgWhite)
			for _, tool := range newTools.Tools {
				prefix.Print("  NEW   ")
				suffix.Printfln(" %s (%s)", tool.Name, tool.Description)
			}

			toolsToShow := updatedInstalledTools
			if updateShowAllTools {
				toolsToShow = updatedTools
			}
			prefix = pterm.NewStyle(pterm.FgBlack, pterm.BgYellow)
			suffix = pterm.NewStyle(pterm.FgWhite)
			for _, tool := range toolsToShow.Tools {
				prefix.Print(" UPDATE ")
				suffix.Printfln(" %s %s", tool.Name, tool.Version)
			}

			if len(newUnigetVersion) > 0 {
				prefix = pterm.NewStyle(pterm.FgBlack, pterm.BgYellow)
				suffix = pterm.NewStyle(pterm.FgWhite)
				prefix.Println()
				prefix.Print("  NEWS  ")
				suffix.Printfln(" Update to uniget %s by running 'uniget self-upgrade'", newUnigetVersion)
			}
		}

		return nil
	},
}

func downloadMetadata() error {
	if metadataDownloaded {
		logging.Debugf("Metadata already downloaded, skipping download")
		return nil
	}

	configuration.AssertCacheDirectory()
	t, err := containers.FindToolRef([]string{constants.Registry}, []string{constants.ImageRepository}, "metadata", constants.MetadataImageTag)
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

	logging.Debugf("Changing directory to %s", configuration.Prefix+"/"+configuration.GetCacheDirectory())
	err = os.Chdir(configuration.Prefix + "/" + configuration.GetCacheDirectory())
	if err != nil {
		return fmt.Errorf("error changing directory to %s: %s", configuration.Prefix+"/"+configuration.GetCacheDirectory(), err)
	}

	progressReader := createProgressReader("Downloading metadata")
	logging.Debugf("Extracting archive to %s", configuration.Prefix+"/"+configuration.GetCacheDirectory())
	err = containers.GetFirstLayerFromRegistry(context.Background(), rc, t.GetRef(), progressReader, func(reader io.ReadCloser) error {
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

	metadataDownloaded = true
	metadataLoaded = false

	return nil
}

func loadMetadata() (err error) {
	if metadataLoaded {
		logging.Debugf("Metadata already loaded, skipping load")
		return nil
	}

	if len(os.Getenv("UNIGET_IGNORE_METADATA_SIGNATURE")) > 0 {
		_, err = security.VerifySigstoreBundle(
			configuration.Prefix+"/"+configuration.GetMetadataFile(),
			configuration.Prefix+"/"+configuration.GetMetadataFile()+".sigstore.json",
			"https://token.actions.githubusercontent.com",
			"",
			"",
			"https://github\\.com/uniget-org/tools/\\.github/workflows/[^.]+\\.yml@refs/heads/main",
		)
		if err != nil {
			return fmt.Errorf("error verifying sigstore bundle for metadata: %s", err)
		}
	}

	tools, err = tool.LoadFromFile(configuration.Prefix + "/" + configuration.GetMetadataFile())
	if err != nil {
		return fmt.Errorf("failed to load metadata from file %s: %s", configuration.Prefix+"/"+configuration.GetMetadataFile(), err)
	}

	metadataLoaded = true

	return nil
}
