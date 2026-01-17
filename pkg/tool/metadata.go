package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/safearchive/tar"
	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gopkg.in/yaml.v3"
)

func NewMetadataFromDirectory(directory string) (Metadata, error) {
	toolNames, err := FindToolsFromFilesystem(directory)
	if err != nil {
		return Metadata{}, fmt.Errorf("error finding tools from filesystem: %s", err)
	}
	return NewMetadataFromToolNames(directory, toolNames)
}

func NewMetadataFromToolNames(directory string, toolNames []string) (Metadata, error) {
	metadata := Metadata{
		Tools: []Tool{},
	}
	for _, toolName := range toolNames {
		tool, err := LoadManifestFromFile(directory + "/" + toolName + "/manifest.yaml")
		if err != nil {
			return Metadata{}, fmt.Errorf("unable to load manifest for %s: %s", toolName, err)
		}
		metadata.Tools = append(metadata.Tools, tool)
	}

	return metadata, nil
}

func NewMetadataFromRegistry(registry string, imageRepository string, metadataImageTag string) (Metadata, error) {
	metadata := Metadata{}

	t, err := containers.FindToolRef([]string{registry}, []string{imageRepository}, "metadata", metadataImageTag)
	if err != nil {
		return metadata, fmt.Errorf("error finding metadata: %s", err)
	}
	rc := containers.GetRegclient()
	defer func() {
		err := rc.Close(context.Background(), t.GetRef())
		if err != nil {
			logging.Warning.Printfln("error closing registry client: %s", err)
		}
	}()

	err = containers.GetFirstLayerFromRegistry(context.Background(), rc, t.GetRef(), func(reader io.ReadCloser) error {
		err = archive.ProcessTarContents(reader, func(reader *tar.Reader, header *tar.Header) error {
			if header.Typeflag == tar.TypeReg && header.Name == "metadata.json" {
				data, err := io.ReadAll(reader)
				if err != nil {
					return fmt.Errorf("error reading metadata.json: %s", err)
				}

				err = json.Unmarshal(data, &metadata)
				if err != nil {
					return fmt.Errorf("error unmarshaling metadata.json: %s", err)
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("error processing tar contents: %s", err)
		}

		return nil
	})
	if err != nil {
		return metadata, fmt.Errorf("error getting first layer from registry: %s", err)
	}

	return metadata, nil
}

func (metadata *Metadata) WriteMetadata(filename string) error {
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return nil
	}
	// #nosec G306 -- This is public data
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %s", filename, err)
	}

	return nil
}

func FindToolsFromFilesystem(directory string) ([]string, error) {
	// TODO: check directory

	c, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("fail to read tools from directory %s: %s", directory, err)
	}

	toolNames := make([]string, 0)
	for _, entry := range c {
		if entry.IsDir() {
			// TODO: check of manifest.yaml exists
			toolNames = append(toolNames, entry.Name())
		}
	}

	return toolNames, nil
}
