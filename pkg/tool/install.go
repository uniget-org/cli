package tool

import (
	"fmt"
	"os"

	"github.com/nicholasdille/docker-setup/pkg/archive"
	"github.com/nicholasdille/docker-setup/pkg/containers"

	"github.com/regclient/regclient/types/blob"
)

func (tool *Tool) Install(registryImagePrefix string, prefix string, altArch string) error {
	err := containers.GetManifest(fmt.Sprintf(registryImagePrefix+"%s:main", tool.Name), altArch, func(blob blob.Reader) error {
		err := os.Chdir(prefix + "/")
		if err != nil {
			return fmt.Errorf("error changing directory to %s: %s", prefix+"/", err)
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
}

func (tool *Tool) Inspect(registryImagePrefix string, altArch string) error {
	err := containers.GetManifest(fmt.Sprintf(registryImagePrefix+"%s:main", tool.Name), altArch, func(blob blob.Reader) error {
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
