package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

var uninstallForce bool

func initUninstallCmd() {
	uninstallCmd.Flags().BoolVar(&uninstallForce, "force", false, "Force uninstallation")

	rootCmd.AddCommand(uninstallCmd)
}

var uninstallCmd = &cobra.Command{
	Use: "uninstall",
	Aliases: []string{
		"u",
	},
	Short:   "Uninstall tool",
	Long:    constants.Header + "\nUninstall tools",
	GroupID: "tool",
	Args:    cobra.OnlyValidArgs,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		configuration.AssertWritableTarget()
		configuration.AssertLibDirectory()

		err = runPreUninstallHooks(args...)
		if err != nil {
			return fmt.Errorf("unable to run pre-uninstall hooks: %s", err)
		}

		for _, toolName := range args {
			tool, err := tools.GetByName(toolName)
			if err != nil {
				return fmt.Errorf("unable to find tool %s: %s", toolName, err)
			}

			err = tool.UpdateStatus(
				configuration.Prefix,
				configuration.Target,
				configuration.GetCacheDirectory(),
				configuration.Arch,
				configuration.AltArch,
			)
			if err != nil {
				return fmt.Errorf("failed to update status for tool %s: %s", tool.Name, err)
			}

			if !uninstallForce && !tool.IsInstalled() {
				logging.Warning.Printfln("Tool %s is not installed", toolName)
				return nil
			}

			var uninstallSpinner *pterm.SpinnerPrinter
			uninstallMessage := fmt.Sprintf("Uninstalling %s", tool.Name)
			if configuration.LogLevel == "warning" {
				uninstallSpinner, _ = pterm.DefaultSpinner.Start(uninstallMessage)
			} else {
				logging.Info.Println(uninstallMessage)
			}

			err = uninstallTool(toolName)
			if err != nil {
				if uninstallSpinner != nil {
					uninstallSpinner.Fail()
				}
				return fmt.Errorf("unable to uninstall tool %s: %s", toolName, err)
			}

			if uninstallSpinner != nil {
				uninstallSpinner.Success()
			}
		}

		err = runPostUninstallHooks(args...)
		if err != nil {
			return fmt.Errorf("unable to run post-uninstall hooks: %s", err)
		}

		return nil
	},
}

func writeInstalledFiles(tool *tool.Tool, installedFiles []string) error {
	fileListDirectory := configuration.GetLibDirectory() + "/manifests"
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

	logging.Tracef("Looking for manifest file for tool %s at %s", tool.Name, configuration.GetLibDirectory()+"/manifests/"+tool.Name+".txt")
	if myos.FileExists(configuration.GetLibDirectory() + "/manifests/" + tool.Name + ".txt") {
		data, err := os.ReadFile(configuration.GetLibDirectory() + "/manifests/" + tool.Name + ".txt")
		if err != nil {
			return fmt.Errorf("unable to read file %s: %s", installFilename, err)
		}
		installedFiles := strings.Split(string(data), "\n")
		err = uninstallFiles(installedFiles)
		if err != nil {
			return fmt.Errorf("unable to uninstall files: %s", err)
		}

	} else {
		logging.Warning.Printfln("Unable to find manifest for %s", tool.Name)
	}

	if myos.DirectoryExists(configuration.GetCacheDirectory() + "/" + tool.Name) {
		entries, err := os.ReadDir(configuration.GetCacheDirectory() + "/" + tool.Name)
		if err != nil {
			return fmt.Errorf("failed to read cache directory for %s: %s", tool.Name, err)
		}
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("unable to get info for %s: %s", info.Name(), err)
			}

			err = os.Remove(configuration.GetCacheDirectory() + "/" + tool.Name + "/" + info.Name())
			if err != nil {
				return fmt.Errorf("unable to remove %s: %s", info.Name(), err)
			}

			if myos.IsDirectoryEmpty(configuration.GetCacheDirectory() + "/" + tool.Name) {
				err = os.Remove(configuration.GetCacheDirectory() + "/" + tool.Name)
				if err != nil {
					return fmt.Errorf("unable to remove empty directory %s: %s", configuration.GetCacheDirectory()+"/"+tool.Name, err)
				}
				logging.Debugf("Removed empty directory %s", configuration.GetCacheDirectory()+"/"+tool.Name)
			}
		}
	}

	if myos.FileExists(configuration.GetLibDirectory() + "/manifests/" + tool.Name + ".json") {
		err = os.Remove(configuration.GetLibDirectory() + "/manifests/" + tool.Name + ".json")
		if err != nil {
			return fmt.Errorf("unable to remove %s: %s", configuration.GetLibDirectory()+"/manifests/"+tool.Name+".json", err)
		}
	}
	if myos.FileExists(configuration.GetLibDirectory() + "/manifests/" + tool.Name + ".txt") {
		err = os.Remove(configuration.GetLibDirectory() + "/manifests/" + tool.Name + ".txt")
		if err != nil {
			return fmt.Errorf("unable to remove %s: %s", configuration.GetLibDirectory()+"/manifests/"+tool.Name+".txt", err)
		}
	}

	err = tool.RemoveMarkerFile(configuration.GetCacheDirectory())
	if os.IsNotExist(err) {
		logging.Debugf("unable to remove marker file because it does not exist")
	} else if err != nil {
		return fmt.Errorf("unable to remove marker file: %s", err)
	}

	return nil
}

func uninstallFiles(installedFiles []string) error {
	logging.Debugf("Working relative to parent directory %s", configuration.Prefix)
	root, err := os.OpenRoot(configuration.Prefix)
	if err != nil {
		return err
	}
	//nolint:errcheck
	defer root.Close()

	for _, file := range installedFiles {
		logging.Debugf("processing %s", file)

		logging.Debugf("stripped line %s", file)
		if file == "" {
			continue
		}

		if strings.HasPrefix(file, "/") {
			if !strings.HasPrefix(file, configuration.Prefix+"/"+configuration.Target) {
				logging.Warning.Printfln("Skipping %s because it is not safe to remove", file)
				continue
			}
		}

		_, err = root.Lstat(file) // #nosec G703 - Path is checked for correct prefix
		if err != nil {
			logging.Debugf("Unable to stat %s: %s", file, err)
			continue
		}

		err = root.Remove(file) // #nosec G703 - Path is checked for correct prefix
		if err != nil {
			return fmt.Errorf("unable to remove %s: %s", file, err)
		}
	}

	return nil
}
