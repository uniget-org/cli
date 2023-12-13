package tool

import (
	"fmt"
	"os"
	"strings"

	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"

	"github.com/pterm/pterm"
	"github.com/regclient/regclient/types/blob"
)

func (tool *Tool) Install(registryImagePrefix string, prefix string, target string, altArch string) error {
	err := containers.GetManifest(fmt.Sprintf(registryImagePrefix+"%s:%s", tool.Name, strings.Replace(tool.Version, "+", "-", -1)), altArch, func(blob blob.Reader) error {
		pterm.Debug.Printfln("Extracting to %s", prefix)
		err := os.Chdir(prefix)
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", prefix, err)
		}
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting working directory")
		}
		pterm.Debug.Printfln("Current directory: %s", dir)
		err = archive.ExtractTarGz(blob, func(path string) string {
			fixedPath := strings.TrimPrefix(path, "usr/local/")
			pterm.Debug.Printfln("fixedPath=%s", fixedPath)
			pterm.Debug.Printfln("          012345678901234567890")
			if len(fixedPath) >= 16 {
				pterm.Debug.Printfln("          %s", fixedPath[0:15])
			}
			if len(fixedPath) >= 16 && fixedPath[0:15] == "var/lib/uniget/" {
				pterm.Debug.Printfln("No need to prepend target")
			} else {
				pterm.Debug.Printfln("Prepending target to %s", fixedPath)
				fixedPath = target + "/" + fixedPath
			}
			return fixedPath
		})
		if err != nil {
			return fmt.Errorf("failed to extract layer: %s", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get manifest: %s", err)
	}

	return nil
}

func (tool *Tool) Inspect(registryImagePrefix string, altArch string) error {
	err := containers.GetManifest(fmt.Sprintf(registryImagePrefix+"%s:%s", tool.Name, strings.Replace(tool.Version, "+", "-", -1)), altArch, func(blob blob.Reader) error {
		result, err := archive.ListTarGz(blob, func(path string) string {
			fixedPath := strings.TrimPrefix(path, "usr/local/")
			return fixedPath
		})
		if err != nil {
			return fmt.Errorf("failed to extract layer: %s", err)
		}

		for _, file := range result {
			fmt.Printf("%s\n", file)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to get manifest: %s", err)
	}

	return nil
}
