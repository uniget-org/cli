package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/logging"
)

func initSelfUpgradeCmd() {
	rootCmd.AddCommand(selfUpgradeCmd)
}

var selfUpgradeCmd = &cobra.Command{
	Use:   "self-upgrade",
	Short: "Self upgrade " + projectName,
	Long:  header + "\nUpgrade " + projectName + " to latest version",
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
		if selfExe != "uniget" {
			pterm.Warning.Printf("Binary must be called uniget but is %s\n", selfExe)
			return nil
		}

		path, err := exec.LookPath(selfExe)
		if err != nil {
			pterm.Error.Printfln("Failed to find %s in PATH", selfExe)
			return fmt.Errorf("failed to find %s in PATH: %s", selfExe, err)
		}
		fmt.Printf("%s is available at %s\n", selfExe, path)
		selfDir := filepath.Dir(path)

		tag := "latest"
		url := fmt.Sprintf("https://github.com/%s/releases/%s/download/uniget_%s_%s.tar.gz", projectRepository, tag, runtime.GOOS, arch)
		logging.Debug.Printfln("Downloading %s", url)
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				re, err := regexp.Compile(`\/uniget-org\/cli\/releases\/download\/(v\d+\.\d+\.\d+)\/`)
				if err != nil {
					return fmt.Errorf("cannot compile regexp: %w", err)
				}

				if re.MatchString(req.URL.Path) {
					tag = re.FindStringSubmatch(req.URL.Path)[1]
				}
				return nil
			},
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %s", err)
		}
		req.Header.Set("Accept", "application/octet-stream")
		req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", projectName, version))
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to download %s: %s", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("failed to download %s: %s", url, resp.Status)
		}

		if fmt.Sprintf("v%s", version) == tag {
			pterm.Info.Printf("Already at latest version %s\n", version)
			return nil
		}

		logging.Debug.Printfln("Extracting tar.gz")
		err = os.Chdir(selfDir)
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", selfDir, err)
		}
		err = os.Remove(selfExe)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %s", selfExe, err)
		}
		err = archive.ExtractTarGz(resp.Body, func(path string) string { return path })
		if err != nil {
			return fmt.Errorf("failed to extract tar.gz: %s", err)
		}

		return nil
	},
}
