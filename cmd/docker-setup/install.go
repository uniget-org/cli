package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	//"github.com/fatih/color"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

var installMode string
var defaultMode bool
var tagsMode bool
var installedMode bool
var check bool
var plan bool
var requestedTools tool.Tools
var plannedTools tool.Tools
var reinstall bool

//var check_mark string = "✓" // Unicode=\u2713 UTF-8=\xE2\x9C\x93 (https://www.compart.com/de/unicode/U+2713)
//var cross_mark string = "✗" // Unicode=\u2717 UTF-8=\xE2\x9C\x97 (https://www.compart.com/de/unicode/U+2717)

func initInstallCmd() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringVarP(&installMode, "mode", "m", "default", "How to install (default, tags, installed)")
	installCmd.Flags().BoolVarP(&defaultMode, "default", "", false, "Install default tools")
	installCmd.Flags().BoolVarP(&tagsMode, "tags", "", false, "Install tool(s) matching tag")
	installCmd.Flags().BoolVarP(&installedMode, "installed", "i", false, "Update installed tool(s)")
	installCmd.Flags().BoolVarP(&plan, "plan", "", false, "Show tool(s) planned installation")
	installCmd.Flags().BoolVarP(&check, "check", "c", false, "Abort after checking versions")
	installCmd.Flags().BoolVarP(&reinstall, "reinstall", "r", false, "Reinstall tool(s)")
	installCmd.MarkFlagsMutuallyExclusive("mode", "default", "tags", "installed")
	installCmd.MarkFlagsMutuallyExclusive("check", "plan")
}

var installCmd = &cobra.Command{
	Use:       "install [tool...]",
	Aliases:   []string{"i"},
	Short:     "Install tools",
	Long:      header + "\nInstall and update tools",
	Args:      cobra.OnlyValidArgs,
	ValidArgs: tools.GetNames(),
	RunE: func(cmd *cobra.Command, args []string) error {

		// Validation checks
		log.Tracef("Found %d argument(s): %+v", len(args), args)
		if installMode == "list" || installMode == "tags" {
			if len(args) == 0 {
				return fmt.Errorf("you must specify at least one tool for mode list or tags")
			}
		}

		assertMetadataFileExists()
		assertMetadataIsLoaded()

		if defaultMode {
			installMode = "default"
		}
		if tagsMode {
			installMode = "tags"
		}
		if installedMode {
			installMode = "only-installed"
		}

		log.Debugf("Using install mode %s", installMode)

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
			log.Debugf("Collecting installed tools")
			for index, tool := range tools.Tools {
				log.Tracef("Getting status for requested tool %s", tool.Name)
				tools.Tools[index].ReplaceVariables(prefix+"/"+target, arch, alt_arch)

				err := tools.Tools[index].GetBinaryStatus()
				if err != nil {
					return fmt.Errorf("unable to determine binary status of %s: %s", tool.Name, err)
				}

				err = tools.Tools[index].GetMarkerFileStatus(cacheDirectory)
				if err != nil {
					return fmt.Errorf("unable to determine marker file status of %s: %s", tool.Name, err)
				}

				if tools.Tools[index].Status.MarkerFilePresent && tools.Tools[index].Status.BinaryPresent {
					log.Tracef("Adding %s to requested tools", tool.Name)
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
				return fmt.Errorf("unable to resolve dependencies for %s: %s", tool.Name, err)
			}
		}
		log.Debugf("Planned %d tool(s)", len(plannedTools.Tools))

		// Populate status of planned tools
		// TODO: Display spinner
		for index, tool := range plannedTools.Tools {
			log.Tracef("Getting status for requested tool %s", tool.Name)
			plannedTools.Tools[index].ReplaceVariables(prefix+"/"+target, arch, alt_arch)

			err := plannedTools.Tools[index].GetBinaryStatus()
			if err != nil {
				return fmt.Errorf("unable to determine binary status of %s: %s", tool.Name, err)
			}

			err = plannedTools.Tools[index].GetMarkerFileStatus(cacheDirectory)
			if err != nil {
				return fmt.Errorf("unable to determine marker file status of %s: %s", tool.Name, err)
			}

			if plannedTools.Tools[index].Status.BinaryPresent && plannedTools.Tools[index].Status.MarkerFilePresent {
				// TODO: Run version check in parallel
				err := plannedTools.Tools[index].GetVersionStatus()
				if err != nil {
					return fmt.Errorf("unable to determine version status of %s: %s", tool.Name, err)
				}
			}
		}

		// Terminate if checking or planning
		if plan || check {
			// TODO: Improve output
			plannedTools.ListWithStatus()
		}
		if check {
			for _, tool := range plannedTools.Tools {
				if !tool.Status.BinaryPresent || !tool.Status.VersionMatches {
					return fmt.Errorf("found missing or outdated tool")
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
			if tool.Status.MarkerFilePresent && !reinstall {
				fmt.Printf("Skipping %s %s because it is already installed.\n", tool.Name, tool.Version)
				continue
			}

			fmt.Printf("%s Installing %s %s", emoji_tool, tool.Name, tool.Version)
			err := tool.Install(registryImagePrefix, prefix+"/", alt_arch)
			fmt.Printf("\n")
			if err != nil {
				return fmt.Errorf("unable to install %s: %s", tool.Name, err)
			}
			tool.CreateMarkerFile(cacheDirectory)
		}

		return nil
	},
}
