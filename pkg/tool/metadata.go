package tool

import (
	"fmt"
	"os"

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

func (metadata *Metadata) WriteMetadata(filename string) error {
	data, err := yaml.Marshal(metadata)
	if err != nil {
		return nil
	}
	err = os.WriteFile(filename, data, 0666)
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
