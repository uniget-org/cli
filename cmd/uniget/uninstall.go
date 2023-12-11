package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/logging"
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

		err = tool.GetBinaryStatus()
		if err != nil {
			return fmt.Errorf("unable to get binary status: %s", err)
		}
		err = tool.GetMarkerFileStatus(prefix + "/" + cacheDirectory)
		if err != nil {
			return fmt.Errorf("unable to get marker file status: %s", err)
		}
		err = tool.GetVersionStatus()
		if err != nil {
			return fmt.Errorf("unable to get version status: %s", err)
		}

		if !force && !tool.Status.MarkerFilePresent && !tool.Status.BinaryPresent {
			pterm.Warning.Printfln("Tool %s is not installed", args[0])
			return nil
		}

		err = uninstallTool(args[0])
		if err != nil {
			return fmt.Errorf("unable to uninstall tool %s: %s", args[0], err)
		}

		logging.Info.Printfln("Uninstalled tool %s", args[0])

		return nil
	},
}

func uninstallTool(toolName string) error {
	tool, err := tools.GetByName(toolName)
	if err != nil {
		return fmt.Errorf("unable to find tool %s: %s", toolName, err)
	}

	if fileExists(prefix + "/" + libDirectory + "/manifests/" + tool.Name + ".txt") {
		data, err := os.ReadFile(prefix + "/" + libDirectory + "/manifests/" + tool.Name + ".txt")
		if err != nil {
			return fmt.Errorf("unable to read file %s: %s", filename, err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			if line == "" {
				continue
			}

			prefixedLine := prefix + "/" + line
			logging.Debug.Printfln("processing %s", prefixedLine)

			_, err := os.Lstat(prefixedLine)
			if err != nil {
				pterm.Debug.Printfln("Unable to stat %s: %s", prefixedLine, err)
				continue
			}

			err = os.Remove(prefixedLine)
			if err != nil {
				return fmt.Errorf("unable to remove %s: %s", prefixedLine, err)
			}
		}
	}

	err = tool.RemoveMarkerFile(prefix + "/" + cacheDirectory)
	if os.IsNotExist(err) {
		logging.Debug.Printfln("unable to remove marker file because it does not exist")
	} else if err != nil {
		return fmt.Errorf("unable to remove marker file: %s", err)
	}

	entries, err := os.ReadDir(prefix + "/" + cacheDirectory + "/" + tool.Name)
	if err != nil {
		return fmt.Errorf("failed to read cache directory for %s: %s", tool.Name, err)
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("unable to get info for %s: %s", info.Name(), err)
		}

		err = os.Remove(prefix + "/" + cacheDirectory + "/" + tool.Name + "/" + info.Name())
		if err != nil {
			return fmt.Errorf("unable to remove %s: %s", info.Name(), err)
		}
	}

	err = os.Remove(prefix + "/" + cacheDirectory + "/" + tool.Name)
	if err != nil {
		return fmt.Errorf("unable to remove %s: %s", prefix+"/"+cacheDirectory+"/"+tool.Name, err)
	}

	if fileExists(prefix + "/" + libDirectory + "/manifests/" + tool.Name + ".json") {
		err = os.Remove(prefix + "/" + libDirectory + "/manifests/" + tool.Name + ".json")
		if err != nil {
			return fmt.Errorf("unable to remove %s: %s", prefix+"/"+libDirectory+"/manifests/"+tool.Name+".json", err)
		}
	}

	return nil
}
