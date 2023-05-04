package tool

import (
	"fmt"
	"os"

	"github.com/nicholasdille/docker-setup/pkg/archive"
	"github.com/nicholasdille/docker-setup/pkg/containers"

	"github.com/regclient/regclient/types/blob"
)

func (tool *Tool) Install(prefix string, alt_arch string) error {
	err := containers.GetManifest(fmt.Sprintf("ghcr.io/nicholasdille/docker-setup/%s:main", tool.Name), alt_arch, func (blob blob.Reader) error {
		os.Chdir(prefix)
		err := archive.ExtractTarGz(blob)
		if err != nil {
			return fmt.Errorf("Failed to extract layer: %s\n", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("Failed to get manifest: %s\n", err)
	}

	return nil
}