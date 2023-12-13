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
		err = archive.ExtractTarGz(blob, func(path string) string {
			return path
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
		result, err := archive.ListTarGz(blob)
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
