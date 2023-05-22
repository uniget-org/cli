package main

import (
	"fmt"

	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	//"github.com/fatih/color"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

var installMode string
var defaultMode bool
var tagsMode bool
var installedMode bool
var check bool
var plan bool
var toolStatus map[string]tool.ToolStatus = make(map[string]tool.ToolStatus)
var requestedTools tool.Tools
var plannedTools tool.Tools
var reinstall bool

//var check_mark string = "✓" // Unicode=\u2713 UTF-8=\xE2\x9C\x93 (https://www.compart.com/de/unicode/U+2713)
//var cross_mark string = "✗" // Unicode=\u2717 UTF-8=\xE2\x9C\x97 (https://www.compart.com/de/unicode/U+2717)

func initInstallCmd() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringVarP(&installMode,   "mode",      "m", "default", "How to install (default, tags, installed)")
	installCmd.Flags().BoolVarP(  &defaultMode,   "default",   "d", false,     "Install default tools")
	installCmd.Flags().BoolVarP(  &tagsMode,      "tags",      "t", false,     "Install tool(s) matching tag")
	installCmd.Flags().BoolVarP(  &installedMode, "installed", "i", false,     "Update installed tool(s)")
	installCmd.Flags().BoolVarP(  &plan,          "plan",      "p", false,     "Show tool(s) planned installation")
	installCmd.Flags().BoolVarP(  &check,         "check",     "c", false,     "Abort after checking versions")
	installCmd.Flags().BoolVarP(  &reinstall,     "reinstall", "r", false,     "Reinstall tool(s)")
	installCmd.MarkFlagsMutuallyExclusive("mode", "default", "tags", "installed")
	installCmd.MarkFlagsMutuallyExclusive("check", "plan")
}

var installCmd = &cobra.Command{
	Use:       "install [tool...]",
	Aliases:   []string{"i"},
	Short:     "Install tools",
	Long:      header + "\nInstall and update tools",
	ValidArgs: tools.GetNames(),
	Args:      cobra.OnlyValidArgs,
	RunE:      func(cmd *cobra.Command, args []string) error {

		// Validation checks
		log.Tracef("Found %d argument(s): %+v", len(args), args)
		if installMode == "list" || installMode == "tags" {
			if len(args) == 0 {
				return fmt.Errorf("You must specify at least one tool for mode list or tags.")
			}
		}

		assertMetadataFileExists()
		tools, err := tool.LoadFromFile(metadataFile)
		if err != nil {
			return fmt.Errorf("Failed to load metadata from file %s: %s\n", metadataFile, err)
		}

		// Fill default values and replace variables
		for index, tool := range tools.Tools {
			log.Tracef("Getting status for requested tool %s", tool.Name)
			tools.Tools[index].ReplaceVariables(prefix + "/" + target, arch, alt_arch)

			err := tools.Tools[index].GetBinaryStatus()
			if err != nil {
				return fmt.Errorf("Unable to determine binary status of %s: %s", tool.Name, err)
			}

			err = tools.Tools[index].GetMarkerFileStatus(cacheDirectory)
			if err != nil {
				return fmt.Errorf("Unable to determine marker file status of %s: %s", tool.Name, err)
			}

			err = tools.Tools[index].GetVersionStatus()
			if err != nil {
				return fmt.Errorf("Unable to determine version status of %s: %s", tool.Name, err)
			}
		}

		if defaultMode {
			installMode = "default"
		}
		if tagsMode {
			installMode = "tags"
		}
		if installedMode {
			installMode = "only-installed"
		}

		// Collect requested tools based on mode
		if len(args) > 0 && installMode == "default" {
			installMode = "list"
		}
		if installMode == "tags" {
			requestedTools = tools.GetByTags(args)

		} else if installMode == "list" {
			requestedTools = tools.GetByNames(args)

		} else if installMode == "default" {
			requestedTools = tools.GetByTags([]string{"category/default"})

		} else if installMode == "only-installed" {
			for _, tool := range tools.Tools {
				if toolStatus[tool.Name].BinaryPresent {
					requestedTools.Tools = append(requestedTools.Tools, tool)
				}
			}
		}
		log.Debugf("Requested %d tool(s)", len(requestedTools.Tools))

		// Add dependencies of requested tools
		// Set installation order
		for _, tool := range requestedTools.Tools {
			err := tools.ResolveDependencies(&plannedTools, tool.Name)
			if err != nil {
				return fmt.Errorf("Unable to resolve dependencies for %s: %s", tool.Name, err)
			}
		}
		log.Debugf("Planned %d tool(s)", len(plannedTools.Tools))

		// Terminate if checking or planning
		if plan || check {
			plannedTools.ListWithStatus(toolStatus)
		}
		if check {
			for _, tool := range plannedTools.Tools {
				if ! toolStatus[tool.Name].BinaryPresent || ! toolStatus[tool.Name].VersionMatches {
					return fmt.Errorf("Found missing or outdated tool")
				}
			}
		}
		if plan || check {
			return nil
		}

		// Install
		assertWritableTarget()
		assertLibDirectory()
		for _, tool := range plannedTools.Tools {
			if tool.Status.MarkerFilePresent && ! reinstall {
				fmt.Printf("Skipping %s %s because it is already installed.\n", tool.Name, tool.Version)
				continue
			}

			fmt.Printf("%s Installing %s %s", emoji_tool, tool.Name, tool.Version)
			err := tool.Install(registryImagePrefix, prefix + "/", alt_arch)
			fmt.Printf("\n")
			if err != nil {
				return fmt.Errorf("Unable to install %s: %s", tool, err)
			}
			tool.CreateMarkerFile(cacheDirectory)
		}

		return nil
	},
}
