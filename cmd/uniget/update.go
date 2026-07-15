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
	myos "gitlab.com/uniget-org/cli/pkg/os"
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
	GroupID: "metadata",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		newRevisionAvailable, err := hasMetadataUpdate()
		if err != nil {
			return fmt.Errorf("error checking for metadata update: %s", err)
		}
		if newRevisionAvailable {
			err = backupMetadata()
			if err != nil {
				return fmt.Errorf("error backing up metadata: %s", err)
			}

			err = downloadMetadata()
			if err != nil {
				err2 := restoreMetadata()
				if err2 != nil {
					logging.Warning.Printfln("Error restoring metadata: %s", err2)
				}
				return fmt.Errorf("error downloading metadata: %s", err)
			}

		} else {
			logging.Info.Println("Metadata is up to date")
		}

		var oldTools *tool.Tools
		oldTools, err = loadMetadata(configuration.Prefix+"/"+configuration.GetMetadataFile()+".bak", true)
		if err != nil {
			return fmt.Errorf("error loading backup metadata: %s", err)
		}
		logging.Debugf("Loaded %d tools from backup", len(oldTools.Tools))
		tools, err = loadMetadata(configuration.Prefix+"/"+configuration.GetMetadataFile(), true)
		if err != nil {
			return fmt.Errorf("error loading metadata: %s", err)
		}
		logging.Debugf("Loaded %d tools from metadata", len(tools.Tools))

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
			for _, tool := range newTools.Tools {
				logging.Customf(pterm.FgBlack, pterm.BgGreen, pterm.FgWhite, pterm.BgDefault, "NEW", " %s (%s)", tool.Name, tool.Description)
			}

			toolsToShow := updatedInstalledTools
			if updateShowAllTools {
				toolsToShow = updatedTools
			}
			for _, tool := range toolsToShow.Tools {
				logging.Customf(pterm.FgBlack, pterm.BgYellow, pterm.FgWhite, pterm.BgDefault, "UPDATE", " %s %s", tool.Name, tool.Version)
			}

			if len(newUnigetVersion) > 0 {
				logging.Customf(pterm.FgBlack, pterm.BgYellow, pterm.FgWhite, pterm.BgDefault, "NEWS", " Update to uniget %s by running 'uniget self-upgrade'", newUnigetVersion)
			}
		}

		return nil
	},
}

func backupMetadata() (err error) {
	metadataFile := configuration.Prefix + "/" + configuration.GetMetadataFile()
	metadataFileSigstore := metadataFile + ".sigstore.json"
	backupMetadataFile := metadataFile + ".bak"
	backupMetadataFileSigstore := backupMetadataFile + ".sigstore.json"

	if myos.FileExists(backupMetadataFile) {
		err = os.Remove(backupMetadataFile)
		if err != nil {
			return fmt.Errorf("error removing old backup metadata file: %s", err)
		}
	}
	if myos.FileExists(backupMetadataFileSigstore) {
		err = os.Remove(backupMetadataFileSigstore)
		if err != nil {
			return fmt.Errorf("error removing old backup metadata sigstore file: %s", err)
		}
	}

	err = myos.CloneFile(metadataFile, backupMetadataFile)
	if err != nil {
		return fmt.Errorf("error backing up metadata file: %s", err)
	}
	err = myos.CloneFile(metadataFileSigstore, backupMetadataFileSigstore)
	if err != nil {
		return fmt.Errorf("error backing up metadata sigstore file: %s", err)
	}

	return nil
}

func restoreMetadata() (err error) {
	metadataFile := configuration.Prefix + "/" + configuration.GetMetadataFile()
	metadataFileSigstore := metadataFile + ".sigstore.json"
	backupMetadataFile := metadataFile + ".bak"
	backupMetadataFileSigstore := backupMetadataFile + ".sigstore.json"

	if myos.FileExists(backupMetadataFile) {
		_ = os.Remove(metadataFile)
		err = os.Rename(backupMetadataFile, metadataFile)
		if err != nil {
			return fmt.Errorf("error restoring backup metadata file: %s", err)
		}
	}
	if myos.FileExists(backupMetadataFileSigstore) {
		_ = os.Remove(metadataFileSigstore)
		err = os.Rename(backupMetadataFileSigstore, metadataFileSigstore)
		if err != nil {
			return fmt.Errorf("error restoring backup metadata sigstore file: %s", err)
		}
	}

	return nil
}

func hasMetadataUpdate() (bool, error) {
	t, err := containers.FindToolRef([]string{constants.Registry}, []string{constants.ImageRepository}, "metadata", constants.MetadataImageTag)
	if err != nil {
		return false, fmt.Errorf("error finding metadata: %s", err)
	}

	labels, err := containers.GetImageLabels(t)
	if err != nil {
		return false, fmt.Errorf("error getting image labels: %s", err)
	}
	if labels["org.opencontainers.image.revision"] == tools.Revision {
		return false, nil
	}

	return true, nil
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

func loadMetadata(filename string, force bool) (loadedTools *tool.Tools, err error) {
	if metadataLoaded && !force {
		logging.Debugf("Metadata already loaded, skipping load")
		return loadedTools, nil
	}

	if len(os.Getenv("UNIGET_IGNORE_METADATA_SIGNATURE")) > 0 {
		_, err = security.VerifySigstoreBundle(
			filename,
			filename+".sigstore.json",
			"https://token.actions.githubusercontent.com",
			"",
			"",
			"https://github\\.com/uniget-org/tools/\\.github/workflows/[^.]+\\.yml@refs/heads/main",
		)
		if err != nil {
			return nil, fmt.Errorf("error verifying sigstore bundle for metadata: %s", err)
		}
	}

	loadedTools, err = tool.LoadFromFile(filename)
	if err != nil {
		return loadedTools, fmt.Errorf("failed to load metadata from file %s: %s", filename, err)
	}

	metadataLoaded = true

	return loadedTools, nil
}
