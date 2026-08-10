package main

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
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
			err = configuration.DownloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}

		} else {
			logging.Info.Println("Metadata is up to date")
		}

		var newTools *tool.Tools
		newTools, err = configuration.LoadMetadata(configuration.GetMetadataFile())
		if err != nil {
			return fmt.Errorf("error loading metadata: %s", err)
		}
		logging.Debugf("Loaded %d tools from metadata", len(newTools.Tools))

		if updateQuiet {
			return nil
		}

		var addedTools tool.Tools
		var updatedTools tool.Tools
		var updatedInstalledTools tool.Tools
		newUnigetVersion := ""
		for _, newTool := range newTools.Tools {
			logging.Debugf("Checking tool %s for updates", newTool.Name)

			if version != "main" && newTool.Name == "uniget" && newTool.Version != version {
				newUnigetVersion = newTool.Version
			}

			tool, err := tools.GetByName(newTool.Name)
			if err != nil {
				addedTools.Tools = append(addedTools.Tools, newTool)
				continue
			}
			err = tool.UpdateStatus(
				configuration.Prefix,
				configuration.Target,
				configuration.GetCacheDirectory(),
				configuration.Arch,
				configuration.AltArch,
			)
			if err != nil {
				logging.Warning.Printfln("Error updating status for %s: %s", tool.Name, err)
			}
			logging.Debugf("  Current version: %s, New version: %s", tool.Status.Version, newTool.Version)
			if tool.IsUpgradable() {
				logging.Debugf("  Tool %s was updated from %s to %s", newTool.Name, tool.Status.Version, newTool.Version)

				updatedTools.Tools = append(updatedTools.Tools, newTool)
				if tool.IsInstalled() {
					logging.Debugf("  Tool %s is installed", newTool.Name)
					updatedInstalledTools.Tools = append(updatedInstalledTools.Tools, newTool)
				}
			}
		}

		logging.Debugf("Showing new tools and updates")
		for _, tool := range addedTools.Tools {
			logging.Customf(pterm.FgBlack, pterm.BgGreen, pterm.FgWhite, pterm.BgDefault, "NEW", " %s (%s)", tool.Name, tool.Description)
		}
		toolsToShow := updatedInstalledTools
		if updateShowAllTools {
			toolsToShow = updatedTools
		}
		for _, tool := range toolsToShow.Tools {
			logging.Customf(pterm.FgBlack, pterm.BgYellow, pterm.FgWhite, pterm.BgDefault, "UPDATE", "%s %s", tool.Name, tool.Version)
		}
		if len(newUnigetVersion) > 0 {
			logging.Customf(pterm.FgBlack, pterm.BgYellow, pterm.FgWhite, pterm.BgDefault, "NEWS", "Update to uniget %s by running 'uniget self-upgrade'", newUnigetVersion)
		}

		return nil
	},
}
