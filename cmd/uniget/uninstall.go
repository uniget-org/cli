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

var force bool

func initUninstallCmd() {
	uninstallCmd.Flags().BoolVar(&force, "force", false, "Force uninstallation")

	rootCmd.AddCommand(uninstallCmd)
}

var uninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Aliases: []string{"u"},
	Short:   "Uninstall tool",
	Long:    header + "\nUninstall tools",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		assertWritableTarget()
		assertLibDirectory()

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("unable to find tool %s: %s", args[0], err)
		}
		checkClientVersionRequirement(tool)

		tool.ReplaceVariables(viper.GetString("prefix")+"/"+viper.GetString("target"), arch, altArch)
		err = tool.GetBinaryStatus()
		if err != nil {
			return fmt.Errorf("unable to get binary status: %s", err)
		}
		err = tool.GetMarkerFileStatus(viper.GetString("prefix") + "/" + cacheDirectory)
		if err != nil {
			return fmt.Errorf("unable to get marker file status: %s", err)
		}
		err = tool.GetVersionStatus()
		if err != nil {
			return fmt.Errorf("unable to get version status: %s", err)
		}

		if !force && !tool.Status.MarkerFilePresent && !tool.Status.BinaryPresent {
			logging.Warning.Printfln("Tool %s is not installed", args[0])
			return nil
		}

		err = uninstallTool(args[0])
		if err != nil {
			return fmt.Errorf("unable to uninstall tool %s: %s", args[0], err)
		}

		return nil
	},
}

func writeInstalledFiles(tool *tool.Tool, installedFiles []string) error {
	fileListDirectory := viper.GetString("prefix") + "/" + libDirectory + "/manifests"
	fileListFilename := fileListDirectory + "/" + tool.Name + ".txt"
	err := os.MkdirAll(fileListDirectory, 0755) // #nosec G301 -- Directory must be accessible by all users
	if err != nil {
		return fmt.Errorf("unable to create directory %s: %s", fileListDirectory, err)
	}

	err = os.WriteFile(fileListFilename, []byte(strings.Join(installedFiles, "\n")), 0644) // #nosec G306 -- File must be world-readable
	if err != nil {
		return fmt.Errorf("unable to open %s: %s", fileListFilename, err)
	}

	return nil
}

func uninstallTool(toolName string) error {
	tool, err := tools.GetByName(toolName)
	if err != nil {
		return fmt.Errorf("unable to find tool %s: %s", toolName, err)
	}

	var uninstallSpinner *pterm.SpinnerPrinter
	installMessage := fmt.Sprintf("Uninstalling %s", tool.Name)
	if viper.GetString("loglevel") == "warning" {
		uninstallSpinner, _ = pterm.DefaultSpinner.Start(installMessage)
	} else {
		logging.Info.Println(installMessage)
	}

	logging.Tracef("Looking for manifest file for tool %s at %s", tool.Name, viper.GetString("prefix")+"/"+libDirectory+"/manifests/"+tool.Name+".txt")
	if fileExists(viper.GetString("prefix") + "/" + libDirectory + "/manifests/" + tool.Name + ".txt") {
		data, err := os.ReadFile(viper.GetString("prefix") + "/" + libDirectory + "/manifests/" + tool.Name + ".txt")
		if err != nil {
			if uninstallSpinner != nil {
				uninstallSpinner.Fail()
			}
			return fmt.Errorf("unable to read file %s: %s", filename, err)
		}
		installedFiles := strings.Split(string(data), "\n")
		err = uninstallFiles(installedFiles)
		if err != nil {
			if uninstallSpinner != nil {
				uninstallSpinner.Fail()
			}
			return fmt.Errorf("unable to uninstall files: %s", err)
		}

	} else {
		logging.Warning.Printfln("Unable to find manifest for %s", tool.Name)
	}

	if directoryExists(viper.GetString("prefix") + "/" + cacheDirectory + "/" + tool.Name) {
		entries, err := os.ReadDir(viper.GetString("prefix") + "/" + cacheDirectory + "/" + tool.Name)
		if err != nil {
			if uninstallSpinner != nil {
				uninstallSpinner.Fail()
			}
			return fmt.Errorf("failed to read cache directory for %s: %s", tool.Name, err)
		}
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				if uninstallSpinner != nil {
					uninstallSpinner.Fail()
				}
				return fmt.Errorf("unable to get info for %s: %s", info.Name(), err)
			}

			err = os.Remove(viper.GetString("prefix") + "/" + cacheDirectory + "/" + tool.Name + "/" + info.Name())
			if err != nil {
				if uninstallSpinner != nil {
					uninstallSpinner.Fail()
				}
				return fmt.Errorf("unable to remove %s: %s", info.Name(), err)
			}
		}

		err = os.Remove(viper.GetString("prefix") + "/" + cacheDirectory + "/" + tool.Name)
		if err != nil {
			if uninstallSpinner != nil {
				uninstallSpinner.Fail()
			}
			return fmt.Errorf("unable to remove %s: %s", viper.GetString("prefix")+"/"+cacheDirectory+"/"+tool.Name, err)
		}
	}

	if fileExists(viper.GetString("prefix") + "/" + libDirectory + "/manifests/" + tool.Name + ".json") {
		err = os.Remove(viper.GetString("prefix") + "/" + libDirectory + "/manifests/" + tool.Name + ".json")
		if err != nil {
			if uninstallSpinner != nil {
				uninstallSpinner.Fail()
			}
			return fmt.Errorf("unable to remove %s: %s", viper.GetString("prefix")+"/"+libDirectory+"/manifests/"+tool.Name+".json", err)
		}
	}
	if fileExists(viper.GetString("prefix") + "/" + libDirectory + "/manifests/" + tool.Name + ".txt") {
		err = os.Remove(viper.GetString("prefix") + "/" + libDirectory + "/manifests/" + tool.Name + ".txt")
		if err != nil {
			if uninstallSpinner != nil {
				uninstallSpinner.Fail()
			}
			return fmt.Errorf("unable to remove %s: %s", viper.GetString("prefix")+"/"+libDirectory+"/manifests/"+tool.Name+".txt", err)
		}
	}

	err = tool.RemoveMarkerFile(viper.GetString("prefix") + "/" + cacheDirectory)
	if os.IsNotExist(err) {
		logging.Debugf("unable to remove marker file because it does not exist")
	} else if err != nil {
		if uninstallSpinner != nil {
			uninstallSpinner.Fail()
		}
		return fmt.Errorf("unable to remove marker file: %s", err)
	}

	if uninstallSpinner != nil {
		uninstallSpinner.Success()
	}

	return nil
}

func uninstallFiles(installedFiles []string) error {
	for _, file := range installedFiles {
		logging.Debugf("processing %s", file)

		logging.Debugf("stripped line %s", file)
		if file == "" {
			continue
		}
		if strings.HasPrefix(file, "/") {
			logging.Warning.Printfln("Skipping %s because it is not safe to remove", file)
			continue
		}

		prefixedFile := viper.GetString("prefix") + "/" + viper.GetString("target") + "/" + file
		logging.Debugf("prefixed line %s", prefixedFile)

		_, err := os.Lstat(prefixedFile)
		if err != nil {
			logging.Debugf("Unable to stat %s: %s", prefixedFile, err)
			continue
		}

		err = os.Remove(prefixedFile)
		if err != nil {
			return fmt.Errorf("unable to remove %s: %s", prefixedFile, err)
		}
	}

	return nil
}
