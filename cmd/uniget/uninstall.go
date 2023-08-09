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
	Use:       "uninstall",
	Aliases:   []string{"u"},
	Short:     "Uninstall tool",
	Long:      header + "\nUninstall tools",
	Args:      cobra.ExactArgs(1),
	ValidArgs: tools.GetNames(),
	RunE: func(cmd *cobra.Command, args []string) error {
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		assertWritableTarget()
		assertLibDirectory()

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("unable to find tool %s: %s", args[0], err)
		}
		tool.GetBinaryStatus()
		tool.GetMarkerFileStatus(prefix + "/" + cacheDirectory)
		tool.GetVersionStatus()
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

			_, err = os.Stat(prefixedLine)
			if err != nil {
				if os.IsNotExist(err) {
					logging.Debug.Printfln("%s does not exist", prefixedLine)
					continue
				}
				return fmt.Errorf("unable to stat %s: %s", prefixedLine, err)
			}

			err = os.Remove(prefixedLine)
			if err != nil {
				pterm.Warning.Printfln("unable to remove %s: %s", prefixedLine, err)
			}
		}
	}

	tool.RemoveMarkerFile(prefix + "/" + cacheDirectory)

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
