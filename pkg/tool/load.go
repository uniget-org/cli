package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/google/safearchive/tar"

	"gitlab.com/uniget-org/cli/pkg/archive"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tui"
)

// isSafePathComponent reports whether value can be used as a single path
// component without escaping the surrounding directory.
func isSafePathComponent(value string) bool {
	if value == "." || !filepath.IsLocal(value) {
		return false
	}
	if value != filepath.Base(value) {
		return false
	}

	return !strings.ContainsAny(value, `/\`)
}

// validateToolPathFields rejects tool names and versions that could escape a
// directory when used as path components, e.g. for marker files, manifests or
// cache directories.
func validateToolPathFields(t *Tool) error {
	if len(t.Name) == 0 {
		return fmt.Errorf("tool name must not be empty")
	}
	if !isSafePathComponent(t.Name) {
		return fmt.Errorf("tool name %q is not a safe path component", t.Name)
	}
	if len(t.Version) > 0 && !isSafePathComponent(t.Version) {
		return fmt.Errorf("version %q of tool %q is not a safe path component", t.Version, t.Name)
	}
	if err := validateToolCheck(t.Check); err != nil {
		return err
	}

	return nil
}

// validateToolCheck rejects version check commands containing control
// characters. Checks are executed through a shell, so multi-line content
// would be equivalent to injecting a script.
func validateToolCheck(check string) error {
	if len(check) == 0 {
		return nil
	}
	if strings.ContainsAny(check, "\x00\n\r") {
		return fmt.Errorf("check %q contains control characters", check)
	}

	return nil
}

func LoadFromFile(filename string) (tools *Tools, err error) {
	data, err := os.ReadFile(filename) // #nosec G304 -- filename is built when LoadFromFile is called
	if err != nil {
		return nil, fmt.Errorf("error loading file contents: %s", err)
	}

	tools, err = LoadFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("error loading data: %s", err)
	}

	return tools, nil
}

func LoadFromReader(data io.ReadCloser) (tools *Tools, err error) {
	tools = &Tools{}
	err = json.NewDecoder(data).Decode(tools)
	if err != nil {
		return nil, err
	}

	for index, tool := range tools.Tools {
		if err := validateToolPathFields(&tool); err != nil {
			return nil, fmt.Errorf("invalid tool metadata: %s", err)
		}

		if tool.Binary == "" {
			tools.Tools[index].Binary = fmt.Sprintf("${target}/bin/%s", tool.Name)
		}
	}

	return tools, nil
}

func LoadFromBytes(data []byte) (*Tools, error) {
	return LoadFromReader(io.NopCloser(bytes.NewReader(data)))
}

func LoadMetadata(registry []string, repository []string, tag string, p tui.ProgressReader) (*Tools, error) {
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

	var metadataJSONReader io.ReadCloser
	err = containers.GetFirstLayerFromRegistry(context.Background(), rc, t.GetRef(), p, func(reader io.ReadCloser) error {
		return archive.ProcessTarContents(reader, func(reader *tar.Reader, header *tar.Header) error {
			if header.Typeflag == tar.TypeReg && header.Name == "metadata.json" {
				metadataJSONReader = io.NopCloser(reader)
			}

			return nil
		})

	})
	if err != nil {
		return nil, fmt.Errorf("error getting first layer from registry: %s", err)
	}

	tools, err := LoadFromReader(metadataJSONReader)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %s", err)
	}

	return tools, nil
}

func LoadMetadataFromRegistry(registry string, imageRepository string, metadataImageTag string, p tui.ProgressReader) ([]byte, error) {
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

	var metadata Tools
	err = containers.GetFirstLayerFromRegistry(context.Background(), rc, t.GetRef(), p, func(reader io.ReadCloser) error {
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
		return nil, fmt.Errorf("error getting first layer from registry: %s", err)
	}

	for index := range metadata.Tools {
		if err := validateToolPathFields(&metadata.Tools[index]); err != nil {
			return nil, fmt.Errorf("invalid tool metadata: %s", err)
		}
	}

	return json.Marshal(metadata)
}

func LoadManifestFromFile(filename string) (Tool, error) {
	data, err := os.ReadFile(filename) // #nosec G304 -- filename is built when LoadFromFile is called
	if err != nil {
		return Tool{}, fmt.Errorf("error loading file contents: %s", err)
	}

	tool, err := LoadManifestFromBytes(data)
	if err != nil {
		return Tool{}, fmt.Errorf("error loading data: %s", err)
	}

	return tool, nil
}

func LoadManifestFromBytes(data []byte) (Tool, error) {
	var tool Tool

	err := yaml.Unmarshal(data, &tool)
	if err != nil {
		return Tool{}, fmt.Errorf("unable to load YAML from byte array: %s", err)
	}

	if err := validateToolPathFields(&tool); err != nil {
		return Tool{}, fmt.Errorf("invalid tool manifest: %s", err)
	}

	return tool, nil
}
