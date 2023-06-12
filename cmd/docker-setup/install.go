package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

var defaultMode bool
var tagsMode bool
var installedMode bool
var check bool
var plan bool
var requestedTools tool.Tools
var plannedTools tool.Tools
var reinstall bool

func initInstallCmd() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().BoolVarP(&defaultMode, "default", "", false, "Install default tools")
	installCmd.Flags().BoolVarP(&tagsMode, "tags", "", false, "Install tool(s) matching tag")
	installCmd.Flags().BoolVarP(&installedMode, "installed", "i", false, "Update installed tool(s)")
	installCmd.Flags().BoolVarP(&plan, "plan", "", false, "Show tool(s) planned installation")
	installCmd.Flags().BoolVarP(&check, "check", "c", false, "Abort after checking versions")
	installCmd.Flags().BoolVarP(&reinstall, "reinstall", "r", false, "Reinstall tool(s)")
	installCmd.MarkFlagsMutuallyExclusive("default", "tags", "installed")
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
		if fileExists(prefix + "/" + metadataFile) {
			log.Tracef("Loaded metadata file from %s", prefix+"/"+metadataFile)
			loadMetadata()
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Introduce --user and adjust libRoot and cacheRoot when set

		assertMetadataFileExists()
		assertMetadataIsLoaded()

		// Collect requested tools based on mode
		// TODO: Add --all flag
		if defaultMode {
			requestedTools = tools.GetByTags([]string{"category/default"})

		} else if tagsMode {
			requestedTools = tools.GetByTags(args)

		} else if installedMode {
			log.Debugf("Collecting installed tools")
			for index, tool := range tools.Tools {
				log.Tracef("Getting status for requested tool %s", tool.Name)
				tools.Tools[index].ReplaceVariables(prefix+"/"+target, arch, alt_arch)

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

		// TODO: Check for conflicts

		// Populate status of planned tools
		// TODO: Display spinner
		for index, tool := range plannedTools.Tools {
			log.Tracef("Getting status for requested tool %s", tool.Name)
			plannedTools.Tools[index].ReplaceVariables(prefix+"/"+target, arch, alt_arch)

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

		// Terminate if checking or planning
		if plan || check {
			// TODO: Improve output
			//       - Show version status (installed, outdated, missing)
			//       - Use color and emoji
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
			if tool.Status.MarkerFilePresent && tool.Status.VersionMatches && !reinstall {
				fmt.Printf("Skipping %s %s because it is already installed.\n", tool.Name, tool.Version)
				continue
			}

			fmt.Printf("%s Installing %s %s", emoji_tool, tool.Name, tool.Version)
			err := tool.Install(registryImagePrefix, prefix+"/", alt_arch)
			fmt.Printf("\n")
			if err != nil {
				return fmt.Errorf("unable to install %s: %s", tool.Name, err)
			}
			tool.CreateMarkerFile(prefix + "/" + cacheDirectory)
		}

		// Call post_install.sh scripts
		if directoryExists(prefix + "/" + libDirectory + "/post_install") {
			entries, err := os.ReadDir(prefix + "/" + libDirectory + "/post_install")
			if err != nil {
				return fmt.Errorf("unable to read post_install directory: %s", err)
			}
			infos := make([]fs.FileInfo, 0, len(entries))
			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					return fmt.Errorf("unable to get info for %s: %s", entry.Name(), err)
				}
				infos = append(infos, info)
			}
			for _, file := range infos {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".sh") {
					fmt.Printf("Running post_install script %s\n", file.Name())

					log.Tracef("Running post_install script %s", prefix+"/"+libDirectory+"/post_install/"+file.Name())
					cmd := exec.Command("/bin/bash", prefix+"/"+libDirectory+"/post_install/"+file.Name())
					cmd.Env = append(os.Environ(),
						"prefix="+prefix,
						"target="+target,
						"arch="+arch,
						"alt_arch="+alt_arch,
					)
					output, err := cmd.CombinedOutput()
					if err != nil {
						return fmt.Errorf("unable to execute post_install script %s: %s", file.Name(), err)
					}
					fmt.Printf("%s\n", output)

					err = os.Remove(prefix + "/" + libDirectory + "/post_install/" + file.Name())
					if err != nil {
						return fmt.Errorf("unable to remove post_install script %s: %s", file.Name(), err)
					}
				}
			}
		}

		return nil
	},
}
