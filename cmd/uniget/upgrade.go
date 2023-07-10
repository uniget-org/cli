package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/nicholasdille/docker-setup/pkg/archive"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
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
		versionRegex := regexp.MustCompile(`^\d+\.\d+\.\d+(-\w+)?$`)
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

		url := fmt.Sprintf("https://github.com/%s/releases/latest/download/docker-setup_%s_%s.tar.gz", repository, runtime.GOOS, arch)
		pterm.Debug.Printfln("Downloading %s", url)
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %s", err)
		}
		req.Header.Set("Accept", "application/octet-stream")
		req.Header.Set("User-Agent", fmt.Sprintf("docker-setup/%s", version))
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to download %s: %s", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("failed to download %s: %s", url, resp.Status)
		}

		pterm.Debug.Printfln("Extracting tar.gz")
		err = os.Chdir(selfDir)
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", selfDir, err)
		}
		err = os.Remove(selfExe)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %s", selfExe, err)
		}
		err = archive.ExtractTarGz(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to extract tar.gz: %s", err)
		}

		return nil
	},
}
