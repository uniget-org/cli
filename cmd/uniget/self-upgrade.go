package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"unicode"

	"github.com/google/safearchive/tar"
	"github.com/spf13/cobra"

	"gitlab.com/uniget-org/cli/internal/common"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

var selfUpgradeVersion = ""
var selfUpgradePath = ""
var selfUpgradeSource = "uniget"

func initSelfUpgradeCmd() {
	var err error

	selfUpgradeCmd.Flags().StringVarP(&selfUpgradeVersion, "version", "v", selfUpgradeVersion, "Version to upgrade to")
	selfUpgradeCmd.Flags().StringVar(&selfUpgradePath, "path", selfUpgradePath, "Binary to upgrade")
	selfUpgradeCmd.Flags().StringVarP(&selfUpgradeSource, "source", "s", selfUpgradeSource, "Source to upgrade from; can either be uniget or release (default: uniget)")

	err = selfUpgradeCmd.Flags().MarkHidden("version")
	if err != nil {
		logging.Error.Printfln("Failed to mark version flag as hidden: %s", err)
	}
	err = selfUpgradeCmd.Flags().MarkHidden("path")
	if err != nil {
		logging.Error.Printfln("Failed to mark path flag as hidden: %s", err)
	}

	rootCmd.AddCommand(selfUpgradeCmd)
}

var selfUpgradeCmd = &cobra.Command{
	Use:     "self-upgrade",
	Aliases: []string{},
	Short:   "Self upgrade " + constants.ProjectName,
	Long:    constants.Header + "\nUpgrade " + constants.ProjectName + " to latest version",
	GroupID: "config",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		unigetTool, err := tools.GetByName("uniget")
		if err != nil {
			return fmt.Errorf("failed to get uniget tool: %s", err)
		}

		requestedVersion := selfUpgradeVersion
		if requestedVersion == "" {
			requestedVersion = unigetTool.Version
		} else {
			unigetTool.Version = requestedVersion
		}

		if requestedVersion == version {
			logging.Info.Printfln("uniget %s is already installed", requestedVersion)
			return nil
		}

		selfExe := filepath.Base(os.Args[0])
		if selfExe == "." {
			return fmt.Errorf("failed to get base name for %s", os.Args[0])
		}
		if selfExe != "uniget" {
			logging.Warning.Printf("Binary must be called uniget but is %s\n", selfExe)
			return nil
		}
		logging.Tracef("Self upgrade binary: %s", selfExe)

		if selfUpgradePath == "" {
			path, err := exec.LookPath(selfExe)
			if err != nil {
				logging.Error.Printfln("Failed to find %s in PATH", selfUpgradePath)
				return fmt.Errorf("failed to find %s in PATH: %s", selfUpgradePath, err)
			}
			logging.Debugf("%s is available at %s\n", selfUpgradePath, path)
			selfUpgradePath = filepath.Dir(path)
		}
		logging.Tracef("Self upgrade path: %s", selfUpgradePath)

		logging.Info.Printfln("Installing version %s", requestedVersion)

		logging.Tracef("Changing directory to %s", selfUpgradePath)
		err = os.Chdir(selfUpgradePath)
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", selfUpgradePath, err)
		}
		logging.Tracef("Removing %s", selfUpgradePath)
		err = os.Remove(selfExe)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %s", selfExe, err)
		}

		switch selfUpgradeSource {
		case "uniget":
			err = selfUpdateFromUniget(unigetTool)
			if err != nil {
				return fmt.Errorf("failed to upgrade from uniget: %s", err)
			}
		case "release":
			err = selfUpdateFromRelease(unigetTool)
			if err != nil {
				return fmt.Errorf("failed to upgrade from release: %s", err)
			}
		default:
			return fmt.Errorf("invalid source %s; must be either uniget or release", selfUpgradeSource)
		}

		return nil
	},
}

func selfUpdateFromUniget(unigetTool *tool.Tool) (err error) {
	registries, repositories := unigetTool.GetSourcesWithFallback(constants.Registry, constants.ImageRepository)
	ref, err := containers.FindToolRef(registries, repositories, unigetTool.Name, unigetTool.Version)
	if err != nil {
		return fmt.Errorf("error finding tool %s:%s: %s", unigetTool.Name, unigetTool.Version, err)
	}
	logging.Debugf("Getting image %s", ref)
	unpackUnigetBinary := func(reader *tar.Reader, header *tar.Header) error {
		logging.Tracef("Processing tar item: %s", header.Name)
		if header.Typeflag == tar.TypeReg && header.Name == "bin/uniget" {
			logging.Debugf("Extracting %s", header.Name)

			err = archive.ExtractFileFromTar(selfUpgradePath, "uniget", reader, header)
			if err != nil {
				return fmt.Errorf("failed to extract %s from tar: %s", header.Name, err)
			}
		}

		return nil
	}

	progressReader := common.CreateProgressReader("Downloading", configuration.Debug || configuration.Trace)
	err = toolCache.Get(ref, progressReader, func(reader io.ReadCloser) error { return nil })
	if err != nil {
		return fmt.Errorf("unable to get image: %s", err)
	}
	err = toolCache.Get(ref, progressReader, func(reader io.ReadCloser) error {
		err := archive.ProcessTarContents(reader, unpackUnigetBinary)
		if err != nil {
			return fmt.Errorf("unable to process tar contents: %s", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to upgrade from image: %s", err)
	}

	return nil
}

func selfUpdateFromRelease(unigetTool *tool.Tool) (err error) {
	url := fmt.Sprintf("https://gitlab.com/%s/cli/-/releases/v%s/downloads/uniget_%s_%s.tar.gz", constants.Organization, unigetTool.Version, string(unicode.ToUpper(rune(runtime.GOOS[0])))+runtime.GOOS[1:], configuration.Arch)
	logging.Debugf("Downloading from %s", url)

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

	err = archive.Gunzip(resp.Body, func(reader io.Reader) error {
		return archive.ProcessTarContents(io.NopCloser(reader), func(tar *tar.Reader, header *tar.Header) error {
			if header.Name == "uniget" {
				logging.Debugf("Extracting %s", header.Name)
				err := archive.CallbackExtractTarItem(tar, header)
				if err != nil {
					return fmt.Errorf("failed to extract %s: %s", header.Name, err)
				}
			}

			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("failed to extract tar.gz: %s", err)
	}

	return nil
}

func downloadReleaseAsset(url string) (*http.Response, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", constants.ProjectName, version))
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %s", url, err)
	}

	return resp, nil
}
