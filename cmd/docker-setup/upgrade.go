package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/nicholasdille/docker-setup/pkg/archive"
	"github.com/nicholasdille/docker-setup/pkg/containers"
	"github.com/regclient/regclient/types/blob"
)

func initUpgradeCmd() {
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade docker-setup",
	Long:  header + "\nUpgrade docker-setup to latest version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		versionRegex := regexp.MustCompile(`\d+\.\d+\.\d+`)
		if !versionRegex.MatchString(version) {
			pterm.Warning.Printf("Version is %s and does not match a.b.c\n", version)
			return nil
		}

		selfExe := filepath.Base(os.Args[0])
		if selfExe == "." {
			return fmt.Errorf("failed to get base name for %s", os.Args[0])
		}
		if selfExe != "docker-setup" {
			pterm.Warning.Printf("Binary must be called docker-setup but is %s\n", selfExe)
			return nil
		}

		selfDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %s", err)
		}
		pterm.Info.Printfln("Replacing docker-setup in %s", selfDir)

		err = containers.GetManifest(fmt.Sprintf("%s%s:main", registryImagePrefix, "docker-setup"), altArch, func(blob blob.Reader) error {
			pterm.Debug.Printfln("Extracting to %s", selfDir)
			err := os.Chdir(selfDir)
			if err != nil {
				return fmt.Errorf("error changing directory to %s: %s", selfDir, err)
			}
			err = archive.ExtractTarGz(blob)
			if err != nil {
				return fmt.Errorf("failed to extract layer: %s", err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to get manifest: %s", err)
		}

		return nil
	},
}
