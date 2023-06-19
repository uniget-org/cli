package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

var defaultMode bool
var tagsMode bool
var installedMode bool
var allMode bool
var skipDependencies bool
var skipConflicts bool
var check bool
var plan bool
var requestedTools tool.Tools
var plannedTools tool.Tools
var reinstall bool

func initInstallCmd() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().BoolVar(&defaultMode, "default", false, "Install default tools")
	installCmd.Flags().BoolVar(&tagsMode, "tags", false, "Install tool(s) matching tag")
	installCmd.Flags().BoolVarP(&installedMode, "installed", "i", false, "Update installed tool(s)")
	installCmd.Flags().BoolVarP(&allMode, "all", "a", false, "Install all tools")
	installCmd.Flags().BoolVar(&plan, "plan", false, "Show tool(s) planned installation")
	installCmd.Flags().BoolVar(&skipDependencies, "skip-deps", false, "Skip dependencies")
	installCmd.Flags().BoolVar(&skipConflicts, "skip-conflicts", false, "Skip conflicting tools")
	installCmd.Flags().BoolVarP(&check, "check", "c", false, "Abort after checking versions")
	installCmd.Flags().BoolVarP(&reinstall, "reinstall", "r", false, "Reinstall tool(s)")
	installCmd.MarkFlagsMutuallyExclusive("default", "tags", "installed", "all")
	installCmd.MarkFlagsMutuallyExclusive("check", "plan")
}

var installCmd = &cobra.Command{
	Use:       "install [tool...]",
	Aliases:   []string{"i"},
	Short:     "Install tools",
	Long:      header + "\nInstall and update tools",
	Args:      cobra.OnlyValidArgs,
	ValidArgs: tools.GetNames(),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return assertLoadMetadata()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		// Collect requested tools based on mode
		if defaultMode {
			requestedTools = tools.GetByTags([]string{"category/default"})

		} else if tagsMode {
			requestedTools = tools.GetByTags(args)

		} else if installedMode {
			log.Debugf("Collecting installed tools")
			for index, tool := range tools.Tools {
				log.Tracef("Getting status for requested tool %s", tool.Name)
				tools.Tools[index].ReplaceVariables(prefix+"/"+target, arch, altArch)

				err := tools.Tools[index].GetBinaryStatus()
				if err != nil {
					return fmt.Errorf("unable to determine binary status of %s: %s", tool.Name, err)
				}

				err = tools.Tools[index].GetMarkerFileStatus(prefix + "/" + cacheDirectory)
				if err != nil {
					return fmt.Errorf("unable to determine marker file status of %s: %s", tool.Name, err)
				}

				if tools.Tools[index].Status.MarkerFilePresent && tools.Tools[index].Status.BinaryPresent {
					log.Tracef("Adding %s to requested tools", tool.Name)
					requestedTools.Tools = append(requestedTools.Tools, tool)
				}
			}

		} else if allMode {
			requestedTools = tools

		} else {
			requestedTools = tools.GetByNames(args)
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
			plannedTools.Tools[index].ReplaceVariables(prefix+"/"+target, arch, altArch)

			err := plannedTools.Tools[index].GetBinaryStatus()
			if err != nil {
				return fmt.Errorf("unable to determine binary status of %s: %s", tool.Name, err)
			}

			err = plannedTools.Tools[index].GetMarkerFileStatus(prefix + "/" + cacheDirectory)
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

		// Check for conflicts
		var conflictsDetected = false
		var conflictsWithInstalled tool.Tools
		var conflictsBetweenPlanned tool.Tools
		for index, tool := range plannedTools.Tools {
			if !tool.Status.BinaryPresent && len(tool.ConflictsWith) > 0 {
				for _, conflict := range tool.ConflictsWith {
					conflictTool, err := plannedTools.GetByName(conflict)
					if err != nil {
						continue
					}
					if plannedTools.Contains(conflict) {
						if conflictTool.Status.BinaryPresent {
							conflictsWithInstalled.Tools = append(conflictsWithInstalled.Tools, tool)
						} else {
							conflictsBetweenPlanned.Tools = append(conflictsBetweenPlanned.Tools, tool)
						}
						conflictsDetected = true

						if skipConflicts {
							plannedTools.Tools[index].Status.SkipDueToConflicts = true
						}
					}
				}
			}
		}
		if conflictsDetected {
			plannedTools.ListWithStatus()
		}
		if len(conflictsWithInstalled.Tools) > 0 {
			log.Errorf("Conflicts with installed tools:")
			for _, conflict := range conflictsWithInstalled.Tools {
				log.Errorf("  %s conflicts with %s", conflict.Name, strings.Join(conflict.ConflictsWith, ", "))
			}
			conflictsDetected = true
		}
		if len(conflictsBetweenPlanned.Tools) > 0 {
			log.Errorf("Conflicts between planned tools:")
			for _, conflict := range conflictsBetweenPlanned.Tools {
				log.Errorf("  %s conflicts with %s", conflict.Name, strings.Join(conflict.ConflictsWith, ", "))
			}
			conflictsDetected = true
		}
		if conflictsDetected && !skipConflicts {
			return fmt.Errorf("conflicts detected")
		}

		// Terminate if checking or planning
		if plan || check {
			// TODO: Improve output
			//       - Show version status (installed, outdated, missing)
			//       - Use color and emoji
			if !conflictsDetected {
				plannedTools.ListWithStatus()
			}
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
			if tool.Status.MarkerFilePresent && tool.Status.VersionMatches && !reinstall {
				fmt.Printf("Skipping %s %s because it is already installed.\n", tool.Name, tool.Version)
				continue
			}
			if tool.Status.SkipDueToConflicts {
				fmt.Printf("Skipping %s because it conflicts with another tool.\n", tool.Name)
				continue
			}
			if skipDependencies && tool.Status.IsDependency {
				fmt.Printf("Skipping %s because it is a dependency (--skip-deps was specified)\n", tool.Name)
				continue
			}

			fmt.Printf("%s Installing %s %s", emojiTool, tool.Name, tool.Version)
			err := tool.Install(registryImagePrefix, prefix+"/", altArch)
			fmt.Printf("\n")
			if err != nil {
				return fmt.Errorf("unable to install %s: %s", tool.Name, err)
			}
			tool.CreateMarkerFile(prefix + "/" + cacheDirectory)
		}

		return postinstall()
	},
}
