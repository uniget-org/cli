package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/uniget-org/cli/pkg/logging"
	"github.com/uniget-org/cli/pkg/tool"
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
var reinstall bool

func initInstallCmd() {
	installCmd.Flags().BoolVar(&defaultMode, "default", false, "Install default tools")
	installCmd.Flags().BoolVar(&tagsMode, "tags", false, "Install tool(s) matching tag")
	installCmd.Flags().BoolVarP(&installedMode, "installed", "i", false, "Update installed tool(s)")
	installCmd.Flags().BoolVarP(&allMode, "all", "a", false, "Install all tools")
	installCmd.Flags().StringVar(&filename, "file", "", "Read tools from file")
	installCmd.Flags().BoolVar(&plan, "plan", false, "Show tool(s) planned installation")
	installCmd.Flags().BoolVar(&skipDependencies, "skip-deps", false, "Skip dependencies")
	installCmd.Flags().BoolVar(&skipConflicts, "skip-conflicts", false, "Skip conflicting tools")
	installCmd.Flags().BoolVar(&check, "check", false, "Abort after checking versions")
	installCmd.Flags().BoolVarP(&reinstall, "reinstall", "r", false, "Reinstall tool(s)")
	installCmd.MarkFlagsMutuallyExclusive("default", "tags", "installed", "all", "file")
	installCmd.MarkFlagsMutuallyExclusive("check", "plan")

	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:     "install [tool...]",
	Aliases: []string{"i"},
	Short:   "Install tools",
	Long:    header + "\nInstall and update tools",
	Args:    cobra.OnlyValidArgs,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("update") {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		var requestedTools tool.Tools

		// Collect requested tools based on mode
		if defaultMode {
			logging.Debugf("Adding default tools to requested tools")
			requestedTools = tools.GetByTags([]string{"category/default"})

		} else if tagsMode {
			logging.Debugf("Adding tools matching tags to requested tools")
			requestedTools = tools.GetByTags(args)

		} else if installedMode {
			logging.Debugf("Collecting installed tools")
			spinnerInstalledTools, _ := pterm.DefaultSpinner.Start("Collecting installed tools...")
			var err error
			requestedTools, err = findInstalledTools(tools)
			if err != nil {
				return fmt.Errorf("unable to find installed tools: %s", err)
			}
			spinnerInstalledTools.Info()

		} else if allMode {
			logging.Debugf("Adding all tools to requested tools")
			requestedTools = tools

		} else if filename != "" {
			logging.Debugf("Adding tools from file %s to requested tools", filename)
			data, err := os.ReadFile(filename) // #nosec G304 -- Accept file from arbitrary location
			if err != nil {
				return fmt.Errorf("unable to read file %s: %s", filename, err)
			}
			for _, line := range strings.Split(string(data), "\n") {
				if len(line) == 0 {
					continue
				} else if strings.HasPrefix(line, "#") {
					continue
				}

				logging.Debugf("Adding %s to requested tools", line)
				tool, err := tools.GetByName(line)
				if err != nil {
					logging.Warning.Printfln("Unable to find tool %s: %s", line, err)
					continue
				}
				requestedTools.Tools = append(requestedTools.Tools, *tool)
			}

		} else {
			logging.Debugf("Adding %s to requested tools", strings.Join(args, ","))
			for _, toolName := range args {
				tool, err := tools.GetByName(toolName)
				if err != nil {
					return fmt.Errorf("unable to find tool %s: %s", toolName, err)
				}
				requestedTools.Tools = append(requestedTools.Tools, *tool)
			}
		}
		logging.Debugf("Requested %d tool(s)", len(requestedTools.Tools))

		return installTools(requestedTools, check, plan, reinstall, skipDependencies, skipConflicts)
	},
}

func findInstalledTools(tools tool.Tools) (tool.Tools, error) {
	var requestedTools tool.Tools
	for index, tool := range tools.Tools {
		logging.Debugf("Getting status for requested tool %s", tool.Name)
		tools.Tools[index].ReplaceVariables(viper.GetString("prefix")+"/"+viper.GetString("target"), arch, altArch)

		err := tools.Tools[index].GetBinaryStatus()
		if err != nil {
			return requestedTools, fmt.Errorf("unable to determine binary status of %s: %s", tool.Name, err)
		}

		err = tools.Tools[index].GetMarkerFileStatus(viper.GetString("prefix") + "/" + cacheDirectory)
		if err != nil {
			return requestedTools, fmt.Errorf("unable to determine marker file status of %s: %s", tool.Name, err)
		}

		if tools.Tools[index].Status.MarkerFilePresent && tools.Tools[index].Status.BinaryPresent {
			logging.Debugf("Adding %s to requested tools", tool.Name)
			requestedTools.Tools = append(requestedTools.Tools, tool)
		}
	}

	return requestedTools, nil
}

func installTools(requestedTools tool.Tools, check bool, plan bool, reinstall bool, skipDependencies bool, skipConflicts bool) error {
	var plannedTools tool.Tools

	// Add dependencies of requested tools
	// Set installation order
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
	logging.Debugf("Planned %d tool(s)", len(plannedTools.Tools))

	// Populate status of planned tools
	for index, tool := range plannedTools.Tools {
		if skipDependencies && !tool.Status.IsRequested {
			continue
		}

		logging.Debugf("Getting status for requested tool %s", tool.Name)

		plannedTools.Tools[index].ReplaceVariables(viper.GetString("prefix")+"/"+viper.GetString("target"), arch, altArch)

		err := plannedTools.Tools[index].GetBinaryStatus()
		if err != nil {
			return fmt.Errorf("unable to determine binary status of %s: %s", tool.Name, err)
		}

		err = plannedTools.Tools[index].GetMarkerFileStatus(viper.GetString("prefix") + "/" + cacheDirectory)
		if err != nil {
			return fmt.Errorf("unable to determine marker file status of %s: %s", tool.Name, err)
		}

		err = plannedTools.Tools[index].GetVersionStatus()
		if err != nil {
			return fmt.Errorf("unable to determine version status of %s: %s", tool.Name, err)
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
		logging.Error.Printfln("Conflicts with installed tools:")
		for _, conflict := range conflictsWithInstalled.Tools {
			logging.Error.Printfln("  %s conflicts with %s", conflict.Name, strings.Join(conflict.ConflictsWith, ", "))
		}
		conflictsDetected = true
	}
	if len(conflictsBetweenPlanned.Tools) > 0 {
		logging.Error.Printfln("Conflicts between planned tools:")
		for _, conflict := range conflictsBetweenPlanned.Tools {
			logging.Error.Printfln("  %s conflicts with %s", conflict.Name, strings.Join(conflict.ConflictsWith, ", "))
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
		if tool.Status.VersionMatches && !reinstall {
			//logging.Skip.Printfln("Skipping %s %s because it is already installed.", tool.Name, tool.Version)
			continue
		}
		if tool.Status.SkipDueToConflicts {
			logging.Skip.Printfln("Skipping %s because it conflicts with another tool.", tool.Name)
			continue
		}
		if skipDependencies && !tool.Status.IsRequested {
			logging.Skip.Printfln("Skipping %s because it is a dependency (--skip-deps was specified)", tool.Name)
			continue
		}

		var installSpinner *pterm.SpinnerPrinter
		if reinstall {
			err := uninstallTool(tool.Name)
			if err != nil {
				logging.Warning.Printfln("Unable to uninstall %s: %s", tool.Name, err)
				continue
			}

		} else if tool.Status.BinaryPresent || tool.Status.MarkerFilePresent {
			err := uninstallTool(tool.Name)
			if err != nil {
				logging.Warning.Printfln("Unable to uninstall %s: %s", tool.Name, err)
				continue
			}
			err = printToolUpdate(tool.Name)
			if err != nil {
				logging.Warning.Printfln("Unable to print tool update: %s", err)
				continue
			}
		}
		installMessage := fmt.Sprintf("Installing %s %s", tool.Name, tool.Version)
		if viper.GetString("loglevel") == "warning" {
			installSpinner, _ = pterm.DefaultSpinner.Start(installMessage)
		} else {
			logging.Info.Println(installMessage)
		}

		if !skipDependencies {
			for _, depName := range tool.RuntimeDependencies {
				dep, err := plannedTools.GetByName(depName)
				if err != nil {
					logging.Error.Printfln("Unable to find dependency %s", depName)
					return fmt.Errorf("unable to find dependency %s", depName)
				}

				err = dep.GetBinaryStatus()
				if err != nil {
					logging.Error.Printfln("Unable to get binary status of dependency %s: %s", depName, err)
					return fmt.Errorf("unable to get binary status of dependency %s: %s", depName, err)
				}
				err = dep.GetMarkerFileStatus(viper.GetString("prefix") + "/" + cacheDirectory)
				if err != nil {
					logging.Error.Printfln("Unable to get marker file status of dependency %s: %s", depName, err)
					return fmt.Errorf("unable to get marker file status of dependency %s: %s", depName, err)
				}
				err = dep.GetVersionStatus()
				if err != nil {
					logging.Error.Printfln("Unable to get version status of dependency %s: %s", depName, err)
					return fmt.Errorf("unable to get version status of dependency %s: %s", depName, err)
				}

				if dep.Status.BinaryPresent || dep.Status.MarkerFilePresent {
					continue
				}
				logging.Error.Printfln("Dependency %s is missing", depName)
				return fmt.Errorf("dependency %s is missing", depName)
			}
		}

		assertDirectory(viper.GetString("prefix") + "/" + viper.GetString("target"))
		var err error
		err = tool.Install(registryImagePrefix, viper.GetString("prefix"), viper.GetString("target"), libDirectory, cacheDirectory)
		if err != nil {
			if installSpinner != nil {
				installSpinner.Fail()
			}
			logging.Warning.Printfln("Unable to install %s: %s", tool.Name, err)
			continue
		}
		if installSpinner != nil {
			installSpinner.Success()
		}

		err = printToolUsage(tool.Name)
		if err != nil {
			logging.Warning.Printfln("Unable to print tool usage: %s", err)
			continue
		}

		err = tool.CreateMarkerFile(viper.GetString("prefix") + "/" + cacheDirectory)
		if err != nil {
			logging.Warning.Printfln("Unable to create marker file: %s", err)
			continue
		}
	}

	if viper.GetBool("user") {
		logging.Warning.Printfln("Post installation is not yet support for user context")
		return nil
	}
	if len(viper.GetString("prefix")) > 0 {
		logging.Warning.Println("Post installation skipped because prefix is set")
		return nil
	}
	err := postinstall()
	if err != nil {
		return fmt.Errorf("unable to run postinstall: %s", err)
	}
	err = installProfileDShim()
	if err != nil {
		return fmt.Errorf("unable to install profile.d shim: %s", err)
	}
	err = installCompletionShim()
	if err != nil {
		return fmt.Errorf("unable to install completion shim: %s", err)
	}
	err = installSystemDUnit()
	if err != nil {
		return fmt.Errorf("unable to install systemd unit: %s", err)
	}

	return nil
}
