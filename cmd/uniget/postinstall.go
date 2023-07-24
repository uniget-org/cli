package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/logging"
)

func initPostinstallCmd() {
	rootCmd.AddCommand(postinstallCmd)
}

var postinstallCmd = &cobra.Command{
	Use:   "postinstall",
	Short: "Run postinstall for tools",
	Long:  header + "\nRun postinstall for tools",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return postinstall()
	},
}

func postinstall() error {
	if directoryExists("/" + libDirectory + "/post_install") {
		entries, err := os.ReadDir("/" + libDirectory + "/post_install")
		if err != nil {
			return fmt.Errorf("unable to read post_install directory: %s", err)
		}
		infos := make([]fs.FileInfo, 0, len(entries))
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("unable to get info for %s: %s", entry.Name(), err)
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".sh") {
				infos = append(infos, info)
			}
		}
		if len(infos) > 0 && len(prefix) > 0 {
			pterm.Warning.Printfln("prefix cannot be set for postinstall scripts to run")
			return nil
		}
		for _, file := range infos {
			logging.Info.Printfln("Running post_install script %s", file.Name())

			logging.Debug.Printfln("Running pre_install script %s", "/"+libDirectory+"/pre_install/"+file.Name())
			cmd := exec.Command("/bin/bash", "/"+libDirectory+"/post_install/"+file.Name())
			cmd.Env = append(os.Environ(),
				"prefix=",
				"target=/"+target,
				"arch="+arch,
				"alt_arch="+altArch,
				"uniget_contrib=/"+libDirectory+"/contrib",
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("unable to execute post_install script %s: %s", file.Name(), err)
			}
			fmt.Printf("%s\n", output)

			err = os.Remove("/" + libDirectory + "/post_install/" + file.Name())
			if err != nil {
				return fmt.Errorf("unable to remove post_install script %s: %s", file.Name(), err)
			}
		}
	}

	return nil
}
