package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/safearchive/tar"
	"github.com/uniget-org/cli/pkg/archive"
	"github.com/uniget-org/cli/pkg/containers"
	"github.com/uniget-org/cli/pkg/logging"
)

func LoadFromFile(filename string) (Tools, error) {
	data, err := os.ReadFile(filename) // #nosec G304 -- filename is built when LoadFromFile is called
	if err != nil {
		return Tools{}, fmt.Errorf("error loading file contents: %s", err)
	}

	tools, err := LoadFromBytes(data)
	if err != nil {
		return Tools{}, fmt.Errorf("error loading data: %s", err)
	}

	return tools, nil
}

func LoadFromBytes(data []byte) (Tools, error) {
	var tools Tools

	err := json.Unmarshal(data, &tools)
	if err != nil {
		return Tools{}, err
	}

	for index, tool := range tools.Tools {
		if tool.Binary == "" {
			tools.Tools[index].Binary = fmt.Sprintf("${target}/bin/%s", tool.Name)
		}

		if tool.SchemaVersion == "" {
			tools.Tools[index].SchemaVersion = "1"
		}
	}

	return tools, nil
}

func LoadMetadataFromRegistry(registry string, imageRepository string, metadataImageTag string) ([]byte, error) {
	t, err := containers.FindToolRef([]string{registry}, []string{imageRepository}, "metadata", metadataImageTag)
	if err != nil {
		return nil, fmt.Errorf("error finding metadata: %s", err)
	}
	rc := containers.GetRegclient()
	defer func() {
		err := rc.Close(context.Background(), t.GetRef())
		if err != nil {
			logging.Warning.Printfln("error closing registry client: %s", err)
		}
	}()

	layer, err := containers.GetFirstLayerFromRegistry(context.Background(), rc, t.GetRef())
	if err != nil {
		return nil, fmt.Errorf("error getting first layer from registry: %s", err)
	}

	var metadata Tools
	err = archive.ProcessTarContents(layer, func(reader *tar.Reader, header *tar.Header) error {
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
		return nil, fmt.Errorf("error processing tar contents: %s", err)
	}

	return json.Marshal(metadata)
}
