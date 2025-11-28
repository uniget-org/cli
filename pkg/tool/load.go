package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/safearchive/tar"

	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/containers"
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

func LoadFromReader(data io.ReadCloser) (Tools, error) {
	var tools Tools

	err := json.NewDecoder(data).Decode(&tools)
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

func LoadFromBytes(data []byte) (Tools, error) {
	return LoadFromReader(io.NopCloser(bytes.NewReader(data)))
}

func LoadMetadata(registry []string, repository []string, tag string) (*Tools, error) {
	t, err := containers.FindToolRef(registry, repository, "metadata", tag)
	if err != nil {
		return nil, fmt.Errorf("error finding metadata: %s", err)
	}
	rc := containers.GetRegclient()
	defer func() {
		err := rc.Close(context.Background(), t.GetRef())
		if err != nil {
			fmt.Printf("Error closing registry client: %s\n", err)
		}
	}()

	var metadataJsonReader io.ReadCloser
	err = containers.GetFirstLayerFromRegistry(context.Background(), rc, t.GetRef(), func(reader io.ReadCloser) error {
		return archive.ProcessTarContents(reader, func(reader *tar.Reader, header *tar.Header) error {
			if header.Typeflag == tar.TypeReg && header.Name == "metadata.json" {
				metadataJsonReader = io.NopCloser(reader)
			}

			return nil
		})

	})
	if err != nil {
		return nil, fmt.Errorf("error getting first layer from registry: %s", err)
	}

	tools, err := LoadFromReader(metadataJsonReader)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %s", err)
	}

	return &tools, nil
}
