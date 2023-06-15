package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initPostinstallCmd() {
	rootCmd.AddCommand(postinstallCmd)
}

var postinstallCmd = &cobra.Command{
	Use:       "install [tool...]",
	Aliases:   []string{"i"},
	Short:     "Install tools",
	Long:      header + "\nInstall and update tools",
	Args:      cobra.OnlyValidArgs,
	ValidArgs: tools.GetNames(),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postinstall()
	},
}

func postinstall() error {
	if len(prefix) > 0 {
		log.Warningf("prefix cannot be set for postinstall scripts to run")
		return nil
	}

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
			infos = append(infos, info)
		}
		for _, file := range infos {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".sh") {
				fmt.Printf("Running post_install script %s\n", file.Name())

				log.Tracef("Running post_install script %s", "/"+libDirectory+"/post_install/"+file.Name())
				cmd := exec.Command("/bin/bash", "/"+libDirectory+"/post_install/"+file.Name())
				cmd.Env = append(os.Environ(),
					"prefix=",
					"target=/"+target,
					"arch="+arch,
					"alt_arch="+altArch,
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
	}

	return nil
}
