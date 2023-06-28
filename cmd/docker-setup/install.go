package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

var defaultMode bool
var tagsMode bool
var installedMode bool
var allMode bool
var filename string
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
	installCmd.Flags().StringVar(&filename, "file", "", "Read tools from file")
	installCmd.Flags().BoolVar(&plan, "plan", false, "Show tool(s) planned installation")
	installCmd.Flags().BoolVar(&skipDependencies, "skip-deps", false, "Skip dependencies")
	installCmd.Flags().BoolVar(&skipConflicts, "skip-conflicts", false, "Skip conflicting tools")
	installCmd.Flags().BoolVarP(&check, "check", "c", false, "Abort after checking versions")
	installCmd.Flags().BoolVarP(&reinstall, "reinstall", "r", false, "Reinstall tool(s)")
	installCmd.MarkFlagsMutuallyExclusive("default", "tags", "installed", "all", "file")
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
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		// Collect requested tools based on mode
		if defaultMode {
			log.Debugf("Adding default tools to requested tools")
			requestedTools = tools.GetByTags([]string{"category/default"})

		} else if tagsMode {
			log.Debugf("Adding tools matching tags to requested tools")
			requestedTools = tools.GetByTags(args)

		} else if installedMode {
			pterm.Debug.Println("Collecting installed tools")
			spinnerInstalledTools, _ := pterm.DefaultSpinner.Start("Collecting installed tools...")
			for index, tool := range tools.Tools {
				pterm.Debug.Printfln("Getting status for requested tool %s", tool.Name)
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
					pterm.Debug.Printfln("Adding %s to requested tools", tool.Name)
					requestedTools.Tools = append(requestedTools.Tools, tool)
				}
			}
			spinnerInstalledTools.Info()

		} else if allMode {
			log.Debugf("Adding all tools to requested tools")
			requestedTools = tools

		} else if filename != "" {
			log.Debugf("Adding tools from file %s to requested tools", filename)
			data, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("unable to read file %s: %s", filename, err)
			}
			for _, line := range strings.Split(string(data), "\n") {
				if len(line) == 0 {
					continue
				} else if strings.HasPrefix(line, "#") {
					continue
				}

				log.Debugf("Adding %s to requested tools", line)
				tool, err := tools.GetByName(line)
				if err != nil {
					pterm.Warning.Printfln("Unable to find tool %s: %s", line, err)
					continue
				}
				requestedTools.Tools = append(requestedTools.Tools, *tool)
			}

		} else {
			log.Debugf("Adding %s to requested tools", strings.Join(args, ","))
			requestedTools = tools.GetByNames(args)
		}
		pterm.Debug.Printfln("Requested %d tool(s)", len(requestedTools.Tools))

		// Add dependencies of requested tools
		// Set installation order
		spinnerResolveDeps, _ := pterm.DefaultSpinner.Start("Resolving dependencies...")
		for _, tool := range requestedTools.Tools {
			err := tools.ResolveDependencies(&plannedTools, tool.Name)
			if err != nil {
				return fmt.Errorf("unable to resolve dependencies for %s: %s", tool.Name, err)
			}
		}
		for _, requestedTool := range requestedTools.Tools {
			tool, err := plannedTools.GetByName(requestedTool.Name)
			if err != nil {
				return fmt.Errorf("unable to find %s in planned tools", requestedTool.Name)
			}
			tool.Status.IsRequested = true
		}
		pterm.Debug.Printfln("Planned %d tool(s)", len(plannedTools.Tools))
		spinnerResolveDeps.Info()

		// Populate status of planned tools
		spinnerGetStatus, _ := pterm.DefaultSpinner.Start("Getting status of requested tools...")
		for index, tool := range plannedTools.Tools {
			if skipDependencies && !tool.Status.IsRequested {
				continue
			}

			pterm.Debug.Printfln("Getting status for requested tool %s", tool.Name)
			plannedTools.Tools[index].ReplaceVariables(prefix+"/"+target, arch, altArch)

			err := plannedTools.Tools[index].GetBinaryStatus()
			if err != nil {
				return fmt.Errorf("unable to determine binary status of %s: %s", tool.Name, err)
			}

			err = plannedTools.Tools[index].GetMarkerFileStatus(prefix + "/" + cacheDirectory)
			if err != nil {
				return fmt.Errorf("unable to determine marker file status of %s: %s", tool.Name, err)
			}

			// TODO: Determine installed version from marker file

			if plannedTools.Tools[index].Status.BinaryPresent {
				// TODO: Run version check in parallel
				err := plannedTools.Tools[index].GetVersionStatus()
				if err != nil {
					return fmt.Errorf("unable to determine version status of %s: %s", tool.Name, err)
				}
			}
		}
		spinnerGetStatus.Info()

		// Check for conflicts
		var conflictsDetected = false
		var conflictsWithInstalled tool.Tools
		var conflictsBetweenPlanned tool.Tools
		spinnerConclicts, _ := pterm.DefaultSpinner.Start("Checking for conflicts...")
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
		spinnerConclicts.Info()
		if conflictsDetected {
			plannedTools.ListWithStatus()
		}
		if len(conflictsWithInstalled.Tools) > 0 {
			pterm.Error.Printfln("Conflicts with installed tools:")
			for _, conflict := range conflictsWithInstalled.Tools {
				pterm.Error.Printfln("  %s conflicts with %s", conflict.Name, strings.Join(conflict.ConflictsWith, ", "))
			}
			conflictsDetected = true
		}
		if len(conflictsBetweenPlanned.Tools) > 0 {
			pterm.Error.Printfln("Conflicts between planned tools:")
			for _, conflict := range conflictsBetweenPlanned.Tools {
				pterm.Error.Printfln("  %s conflicts with %s", conflict.Name, strings.Join(conflict.ConflictsWith, ", "))
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
			if (tool.Status.MarkerFilePresent || tool.Status.VersionMatches) && !reinstall {
				pterm.Info.Printfln("Skipping %s %s because it is already installed.", tool.Name, tool.Version)
				continue
			}
			if tool.Status.SkipDueToConflicts {
				pterm.Info.Printfln("Skipping %s because it conflicts with another tool.", tool.Name)
				continue
			}
			if skipDependencies && !tool.Status.IsRequested {
				pterm.Info.Printfln("Skipping %s because it is a dependency (--skip-deps was specified)", tool.Name)
				continue
			}

			fmt.Printf("%s Installing %s %s\n", emojiTool, tool.Name, tool.Version)

			if !skipDependencies {
				for _, toolName := range tool.RuntimeDependencies {
					tool, err := plannedTools.GetByName(toolName)
					if err != nil {
						pterm.Error.Printfln("Unable to find dependency %s", toolName)
						return fmt.Errorf("unable to find dependency %s", toolName)
					}
					tool.GetBinaryStatus()
					if tool.Status.BinaryPresent {
						continue
					}
					pterm.Error.Printfln("Dependency %s is missing", toolName)
					return fmt.Errorf("dependency %s is missing", toolName)
				}
			}

			err := tool.Install(registryImagePrefix, prefix+"/", altArch)
			if err != nil {
				pterm.Warning.Printfln("Unable to install %s: %s", tool.Name, err)
				continue
			}
			// TODO: Remove all marker files
			tool.CreateMarkerFile(prefix + "/" + cacheDirectory)
		}

		return postinstall()
	},
}
