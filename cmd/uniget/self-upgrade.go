package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"

	"github.com/google/go-github/github"
	"github.com/google/safearchive/tar"
	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/logging"
)

var requestedVersion string
var allowPrereleaseVersion bool
var dryRun bool

func initSelfUpgradeCmd() {
	rootCmd.AddCommand(selfUpgradeCmd)

	selfUpgradeCmd.Flags().StringVar(&requestedVersion, "version", "latest", "Upgrade to a specific version")
	selfUpgradeCmd.Flags().BoolVar(&allowPrereleaseVersion, "allow-prerelease", false, "Allow upgrading to prerelease version")
	selfUpgradeCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Do not perform the upgrade, only show what would be done")
}

var selfUpgradeCmd = &cobra.Command{
	Use:   "self-upgrade",
	Short: "Self upgrade " + projectName,
	Long:  header + "\nUpgrade " + projectName + " to latest version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		versionRegex, err := regexp.Compile(`^\d+\.\d+\.\d+(-[a-z]+\.\d+)?$`)
		if err != nil {
			return fmt.Errorf("cannot compile regexp: %w", err)
		}
		if !versionRegex.MatchString(version) {
			return fmt.Errorf("invalid version %s", version)
		}

		selfExe := filepath.Base(os.Args[0])
		if selfExe == "." {
			return fmt.Errorf("failed to get base name for %s", os.Args[0])
		}
		if selfExe != "uniget" {
			logging.Warning.Printf("Binary must be called uniget but is %s\n", selfExe)
			return nil
		}

		path, err := exec.LookPath(selfExe)
		if err != nil {
			logging.Error.Printfln("Failed to find %s in PATH", selfExe)
			return fmt.Errorf("failed to find %s in PATH: %s", selfExe, err)
		}
		logging.Debugf("%s is available at %s\n", selfExe, path)
		selfDir := filepath.Dir(path)

		if allowPrereleaseVersion && requestedVersion == "latest" {
			logging.Debugf("Allowing prerelease version")

			githubClient := github.NewClient(nil)
			releases, _, err := githubClient.Repositories.ListReleases(context.Background(), githubOrganization, "cli", &github.ListOptions{PerPage: 100})
			if err != nil {
				return fmt.Errorf("failed to list releases: %s", err)
			}

			versions := make([]*goversion.Version, 0)
			for _, release := range releases {
				version, err := goversion.NewSemver(*release.TagName)
				if err != nil {
					continue
				}

				versions = append(versions, version)
				sort.Sort(goversion.Collection(versions))
			}
			requestedVersion = versions[len(versions)-1].String()
		}

		var url string
		if requestedVersion == "latest" {
			url = fmt.Sprintf("https://github.com/%s/releases/%s/download/uniget_%s_%s.tar.gz", projectRepository, requestedVersion, runtime.GOOS, arch)
		} else {
			logging.Info.Printfln("Requested version %s", requestedVersion)
			url = fmt.Sprintf("https://github.com/%s/releases/download/v%s/uniget_%s_%s.tar.gz", projectRepository, requestedVersion, runtime.GOOS, arch)
		}

		logging.Debugf("Downloading from %s", url)
		if dryRun {
			logging.Info.Printfln("Would download version %s from %s", requestedVersion, url)
			return nil
		}

		resp, err := downloadReleaseAsset(url)
		if err != nil {
			return fmt.Errorf("failed to download %s: %s", url, err)
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				logging.Warning.Printfln("failed to close response body: %s", err)
			}
		}()

		if resp.StatusCode != 200 {
			return fmt.Errorf("failed to download %s: %s", url, resp.Status)
		}

		requestedVersionVersion, err := goversion.NewVersion(requestedVersion)
		if err != nil {
			return fmt.Errorf("failed to parse version %s: %s", requestedVersion, err)
		}
		versionVersion, err := goversion.NewVersion(version)
		if err != nil {
			return fmt.Errorf("failed to parse current version %s: %s", version, err)
		}
		if requestedVersionVersion.LessThanOrEqual(versionVersion) {
			logging.Info.Printfln("Latest version %s already installed.", version)
			return nil
		}

		logging.Debugf("Extracting tar.gz")
		err = os.Chdir(selfDir)
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", selfDir, err)
		}
		err = os.Remove(selfExe)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %s", selfExe, err)
		}

		bodyGz, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %s", err)
		}
		body, err := archive.Gunzip(bodyGz)
		if err != nil {
			return fmt.Errorf("failed to gunzip body: %s", err)
		}
		err = archive.ProcessTarContents(body, func(tar *tar.Reader, header *tar.Header) error {
			if header.Name == "uniget" {
				logging.Debugf("Extracting %s", header.Name)
				err := archive.CallbackExtractTarItem(tar, header)
				if err != nil {
					return fmt.Errorf("failed to extract %s: %s", header.Name, err)
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to extract tar.gz: %s", err)
		}

		logging.Info.Printfln("Upgraded to version %s", requestedVersion)
		return nil
	},
}

func downloadReleaseAsset(url string) (*http.Response, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			re, err := regexp.Compile(`\/uniget-org\/cli\/releases\/download\/(v\d+\.\d+\.\d+)\/`)
			if err != nil {
				return fmt.Errorf("cannot compile regexp: %w", err)
			}

			if re.MatchString(req.URL.Path) {
				requestedVersion = re.FindStringSubmatch(req.URL.Path)[1]
			}
			return nil
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", projectName, version))
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %s", url, err)
	}
	return resp, nil
}
