package config

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/google/safearchive/tar"

	"gitlab.com/uniget-org/cli/internal/common"
	"gitlab.com/uniget-org/cli/internal/constants"
	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/security"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

func (c *Config) HasMetadataUpdate(revision string) (bool, error) {
	t, err := containers.FindToolRef([]string{constants.Registry}, []string{constants.ImageRepository}, "metadata", constants.MetadataImageTag)
	if err != nil {
		return false, fmt.Errorf("error finding metadata: %s", err)
	}

	labels, err := containers.GetImageLabels(t)
	if err != nil {
		return false, fmt.Errorf("error getting image labels: %s", err)
	}
	if labels["org.opencontainers.image.revision"] == revision {
		return false, nil
	}

	return true, nil
}

func (c *Config) DownloadMetadata() error {
	c.AssertCacheDirectory()
	t, err := containers.FindToolRef([]string{constants.Registry}, []string{constants.ImageRepository}, "metadata", constants.MetadataImageTag)
	if err != nil {
		return fmt.Errorf("error finding metadata: %s", err)
	}
	rc := containers.GetRegclient()
	defer func() {
		err := rc.Close(context.Background(), t.GetRef())
		if err != nil {
			logging.Warning.Printfln("error closing registry client: %s", err)
		}
	}()

	logging.Debugf("Changing directory to %s", c.GetCacheDirectory())
	err = os.Chdir(c.GetCacheDirectory())
	if err != nil {
		return fmt.Errorf("error changing directory to %s: %s", c.GetCacheDirectory(), err)
	}

	progressReader := common.CreateProgressReader("Downloading metadata", c.Debug || c.Trace)
	logging.Debugf("Extracting archive to %s", c.GetCacheDirectory())
	err = containers.GetFirstLayerFromRegistry(context.Background(), rc, t.GetRef(), progressReader, func(reader io.ReadCloser) error {
		err := archive.ProcessTarContents(reader, func(reader *tar.Reader, header *tar.Header) error {
			err := archive.CallbackExtractTarItem(reader, header)
			if err != nil {
				return fmt.Errorf("error extracting tar item: %s", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error processing tar contents: %s", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error getting first layer from registry: %s", err)
	}

	return nil
}

func (c *Config) LoadMetadata(filename string) (loadedTools *tool.Tools, err error) {
	if len(os.Getenv("UNIGET_IGNORE_METADATA_SIGNATURE")) > 0 {
		_, err = security.VerifySigstoreBundle(
			filename,
			filename+".sigstore.json",
			"https://token.actions.githubusercontent.com",
			"",
			"",
			"https://github\\.com/uniget-org/tools/\\.github/workflows/[^.]+\\.yml@refs/heads/main",
		)
		if err != nil {
			return nil, fmt.Errorf("error verifying sigstore bundle for metadata: %s", err)
		}
	}

	loadedTools, err = tool.LoadFromFile(filename)
	if err != nil {
		return loadedTools, fmt.Errorf("failed to load metadata from file %s: %s", filename, err)
	}

	return loadedTools, nil
}
