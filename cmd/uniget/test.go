package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

func initTestCmd() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:     "test",
	Aliases: []string{},
	Short:   "Test stuff",
	Long:    header + "\nTest stuff",
	Hidden:  true,
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

		conflictsDetected := false
		preInstallHookTools := make([]string, 0)
		for _, toolName := range args {
			tools.TraverseRuntimeDependencies(toolName, skipDependencies, func(requestedTool *tool.Tool) error {
				requestedTool.UpdateStatus(viper.GetString("prefix"), viper.GetString("target"), cacheDirectory, arch, altArch)
				requestedTool.Status.IsRequested = true

				if !requestedTool.IsInstalled() {
					for _, conflictedName := range requestedTool.ConflictsWith {
						conflictedTool, err := tools.GetByName(conflictedName)
						if err != nil {
							return fmt.Errorf("unable to find tool %s: %s", conflictedTool.Name, err)
						}
						conflictedTool.UpdateStatus(viper.GetString("prefix"), viper.GetString("target"), cacheDirectory, arch, altArch)

						if conflictedTool.IsInstalled() {
							logging.Warning.Printfln("%s: Conflicted with %s detected.", requestedTool.Name, conflictedTool.Name)
							conflictsDetected = true
						}
					}
				}

				preInstallHookTools = append(preInstallHookTools, requestedTool.Name)

				return nil
			})
		}
		if conflictsDetected {
			return fmt.Errorf("conflicts detected")
		}

		// TODO: plan
		// TODO: check

		err := runPreInstallHooks(preInstallHookTools...)
		if err != nil {
			return fmt.Errorf("unable to run pre-install hooks: %s", err)
		}

		postInstallHookTools := make([]string, 0)
		for _, toolName := range args {
			tools.TraverseRuntimeDependencies(toolName, skipDependencies, func(requestedTool *tool.Tool) error {

				uninstall := false
				var installSpinner *pterm.SpinnerPrinter
				installMessage := fmt.Sprintf("Installing %s %s", requestedTool.Name, requestedTool.Version)
				if reinstall {
					installMessage = fmt.Sprintf("Reinstalling %s %s", requestedTool.Name, requestedTool.Version)

				} else if requestedTool.IsInstalled() {
					uninstall = true
					installMessage = fmt.Sprintf("Updating %s %s", requestedTool.Name, requestedTool.Version)
				}
				installSpinner, _ = pterm.DefaultSpinner.Start(installMessage)
				if uninstall {
					err := uninstallTool(requestedTool.Name)
					if err != nil {
						logging.Warning.Printfln("Unable to uninstall %s: %s", requestedTool.Name, err)
						return nil
					}
					err = printToolUpdate(cmd.OutOrStdout(), requestedTool.Name)
					if err != nil {
						logging.Warning.Printfln("Unable to print tool update: %s", err)
						return nil
					}
				}

				requestedTool.UpdateStatus(viper.GetString("prefix"), viper.GetString("target"), cacheDirectory, arch, altArch)
				requestedTool.Status.IsRequested = true

				assertDirectory(viper.GetString("prefix") + "/" + viper.GetString("target"))
				// Change working directory to prefix
				// so that unpacking can ignore the target directory
				installDir := viper.GetString("prefix")
				if len(installDir) == 0 {
					installDir = "/"
				}
				err := os.Chdir(installDir)
				if err != nil {
					installSpinner.Fail()
					return fmt.Errorf("error changing directory to %s: %s", installDir, err)
				}
				dir, err := os.Getwd()
				if err != nil {
					installSpinner.Fail()
					return fmt.Errorf("error getting working directory")
				}
				logging.Debugf("Current directory: %s", dir)

				var pathToTar string
				var layer io.ReadCloser
				var installedFiles []string
				installTool := func(plannedTool tool.Tool, layer io.ReadCloser) error {
					installedFiles, err = plannedTool.Install(cmd.OutOrStdout(), layer, pathRewriteRules, createPatchFileCallback(plannedTool))
					if err != nil {
						logging.Error.Printfln("Unable to install %s: %s", plannedTool.Name, err)
						logging.Warning.Printfln("Removing partial installation")
						err = uninstallFiles(installedFiles)
						if err != nil {
							logging.Warning.Printfln("Unable to remove partial installation: %s", err)
						}
						return fmt.Errorf("unable to install %s: %s", plannedTool.Name, err)
					}

					return nil
				}
				pathToTar, ok := pathToTarMappings[requestedTool.Name]
				installSuccessful := true
				if ok {
					logging.Debugf("Using tar file mappings for installation")
					if _, err := os.Stat(pathToTar); os.IsNotExist(err) {
						installSpinner.Fail()
						return fmt.Errorf("tar file %s does not exist", pathToTar)
					}
					layer, err = os.Open(pathToTar) // #nosec G304 -- Location supplied by user
					if err != nil {
						installSpinner.Fail()
						return fmt.Errorf("unable to read tar file %s: %s", pathToTar, err)
					}
					//nolint:errcheck
					defer layer.Close()
					err = installTool(*requestedTool, layer)
					if err != nil {
						installSuccessful = false
					}

				} else {
					logging.Debugf("Using default behaviour for installation")
					registries, repositories := requestedTool.GetSourcesWithFallback(registry, imageRepository)
					ref, err := containers.FindToolRef(registries, repositories, requestedTool.Name, "main")
					if err != nil {
						installSpinner.Fail()
						return fmt.Errorf("error finding tool %s:%s: %s", requestedTool.Name, requestedTool.Version, err)
					}
					logging.Debugf("Getting image %s", ref)
					err = toolCache.Get(ref, func(reader io.ReadCloser) error { return nil })
					if err != nil {
						installSpinner.Fail()
						return fmt.Errorf("unable to get image: %s", err)
					}
					err = toolCache.Get(ref, func(reader io.ReadCloser) error {
						err := installTool(*requestedTool, reader)
						if err != nil {
							installSuccessful = false
							installSpinner.Fail()
							return fmt.Errorf("unable to install %s: %s", requestedTool.Name, err)
						}
						return nil
					})
					if err != nil {
						installSpinner.Fail()
						return fmt.Errorf("unable to install from image: %s", err)
					}
				}
				if !installSuccessful {
					installSpinner.Fail()
					return nil
				}

				logging.Debugf("Installed files: %d", len(installedFiles))
				logging.Tracef("Installed files: %v", installedFiles)
				err = writeInstalledFiles(requestedTool, installedFiles)
				if err != nil {
					logging.Error.Printfln("Unable to write installed files: %s", err)
				}
				requestedToolJson, err := json.MarshalIndent(requestedTool, "", "  ")
				if err != nil {
					logging.Error.Printfln("Unable to marshal tool: %s", err)
				}
				manifestFilename := viper.GetString("prefix") + "/" + libDirectory + "/manifests/" + requestedTool.Name + ".json"
				err = os.WriteFile(manifestFilename, []byte(requestedToolJson), 0644) // #nosec G306 -- File must be world-readable
				if err != nil {
					logging.Error.Printfln("Unable to write manifest file: %s", err)
				}
				installSpinner.Success()

				err = printToolUsage(cmd.OutOrStdout(), requestedTool.Name)
				if err != nil {
					logging.Warning.Printfln("Unable to print tool usage: %s", err)
					return nil
				}

				err = requestedTool.CreateMarkerFile(viper.GetString("prefix") + "/" + cacheDirectory)
				if err != nil {
					logging.Warning.Printfln("Unable to create marker file: %s", err)
					return nil
				}

				postInstallHookTools = append(postInstallHookTools, requestedTool.Name)

				return nil
			})
		}

		err = installProfileDShim()
		if err != nil {
			return fmt.Errorf("unable to install profile.d shim: %s", err)
		}

		if len(postInstallHookTools) > 0 {
			err = runPostInstallHooks(postInstallHookTools...)
			if err != nil {
				return fmt.Errorf("unable to run post-install hooks: %s", err)
			}
		}

		return nil
	},
}
