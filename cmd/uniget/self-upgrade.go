package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/safearchive/tar"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
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
		err := downloadMetadata()
		if err != nil {
			return fmt.Errorf("error downloading metadata: %s", err)
		}
		assertMetadataFileExists()
		assertMetadataIsLoaded()
		err = loadMetadata()
		if err != nil {
			return fmt.Errorf("error loading metadata: %s", err)
		}

		unigetTool, err := tools.GetByName("uniget")
		if err != nil {
			return fmt.Errorf("failed to get uniget tool: %s", err)
		}

		if unigetTool.Version == version {
			logging.Info.Printfln("uniget %s is already installed", unigetTool.Version)
			return nil
		}

		var installSpinner *pterm.SpinnerPrinter
		installSpinner, _ = pterm.DefaultSpinner.Start("Upgrading uniget")

		selfExe := filepath.Base(os.Args[0])
		if selfExe == "." {
			installSpinner.Fail()
			return fmt.Errorf("failed to get base name for %s", os.Args[0])
		}
		if selfExe != "uniget" {
			installSpinner.Fail()
			logging.Warning.Printf("Binary must be called uniget but is %s\n", selfExe)
			return nil
		}

		path, err := exec.LookPath(selfExe)
		if err != nil {
			installSpinner.Fail()
			logging.Error.Printfln("Failed to find %s in PATH", selfExe)
			return fmt.Errorf("failed to find %s in PATH: %s", selfExe, err)
		}
		logging.Debugf("%s is available at %s\n", selfExe, path)
		selfDir := filepath.Dir(path)

		logging.Info.Printfln("Installing version %s", unigetTool.Version)

		logging.Tracef("Changing directory to %s", selfDir)
		err = os.Chdir(selfDir)
		if err != nil {
			installSpinner.Fail()
			return fmt.Errorf("error changing directory to %s: %s", selfDir, err)
		}
		logging.Tracef("Removing %s", selfExe)
		err = os.Remove(selfExe)
		if err != nil {
			installSpinner.Fail()
			return fmt.Errorf("failed to remove %s: %s", selfExe, err)
		}

		registries, repositories := unigetTool.GetSourcesWithFallback(registry, imageRepository)
		ref, err := containers.FindToolRef(registries, repositories, unigetTool.Name, "main")
		if err != nil {
			installSpinner.Fail()
			return fmt.Errorf("error finding tool %s:%s: %s", unigetTool.Name, unigetTool.Version, err)
		}
		logging.Debugf("Getting image %s", ref)
		unpackUnigetBinary := func(reader *tar.Reader, header *tar.Header) error {
			logging.Tracef("Processing tar item: %s", header.Name)
			if header.Typeflag == tar.TypeReg && header.Name == "bin/uniget" {
				logging.Debugf("Extracting %s", header.Name)

				err = archive.ExtractFileFromTar(selfDir, "uniget", reader, header)
				if err != nil {
					return fmt.Errorf("failed to extract %s from tar: %s", header.Name, err)
				}
			}

			return nil
		}
		err = toolCache.Get(ref, func(reader io.ReadCloser) error { return nil })
		if err != nil {
			installSpinner.Fail()
			return fmt.Errorf("unable to get image: %s", err)
		}
		err = toolCache.Get(ref, func(reader io.ReadCloser) error {
			err := archive.ProcessTarContents(reader, unpackUnigetBinary)
			if err != nil {
				return fmt.Errorf("unable to process tar contents: %s", err)
			}

			return nil
		})
		if err != nil {
			installSpinner.Fail()
			return fmt.Errorf("unable to upgrade from image: %s", err)
		}

		installSpinner.Success()
		return nil
	},
}
