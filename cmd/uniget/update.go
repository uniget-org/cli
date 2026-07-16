package main

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
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
		newRevisionAvailable, err := configuration.HasMetadataUpdate(tools.Revision)
		if err != nil {
			return fmt.Errorf("error checking for metadata update: %s", err)
		}
		if newRevisionAvailable {
			err = configuration.BackupMetadata()
			if err != nil {
				return fmt.Errorf("error backing up metadata: %s", err)
			}

			err = configuration.DownloadMetadata()
			if err != nil {
				err2 := configuration.RestoreMetadata()
				if err2 != nil {
					logging.Warning.Printfln("Error restoring metadata: %s", err2)
				}
				return fmt.Errorf("error downloading metadata: %s", err)
			}

		} else {
			logging.Info.Println("Metadata is up to date")
		}

		var oldTools *tool.Tools
		if myos.FileExists(configuration.Prefix + "/" + configuration.GetMetadataFile() + ".bak") {
			oldTools, err = configuration.LoadMetadata(configuration.Prefix + "/" + configuration.GetMetadataFile() + ".bak")
			if err != nil {
				return fmt.Errorf("error loading backup metadata: %s", err)
			}
			logging.Debugf("Loaded %d tools from backup", len(oldTools.Tools))

		} else {
			oldTools = tools
		}

		tools, err = configuration.LoadMetadata(configuration.Prefix + "/" + configuration.GetMetadataFile())
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
