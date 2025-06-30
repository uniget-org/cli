package main

import (
	"fmt"
	"os"

	"github.com/uniget-org/cli/pkg/tool"
	"gopkg.in/yaml.v2"
)

type UnigetTool struct {
	Subdirectory string
	FullPath     string
	Tool         tool.Tool
}

type UnigetTools struct {
	BaseDirectory     string
	ToolsSubDirectory string
	Directory         string
	Tools             map[string]UnigetTool
}

func NewUnigetTools(baseDirectory string) *UnigetTools {
	toolsSubdirectory := "tools"
	return &UnigetTools{
		BaseDirectory:     baseDirectory,
		ToolsSubDirectory: toolsSubdirectory,
		Directory:         fmt.Sprintf("%s/%s", baseDirectory, toolsSubdirectory),
		Tools:             make(map[string]UnigetTool),
	}
}

func NetUnigetTool(directory string, subdirectory string) *UnigetTool {
	return &UnigetTool{
		Subdirectory: subdirectory,
		FullPath:     fmt.Sprintf("%s/%s", directory, subdirectory),
	}
}

func (t *UnigetTool) HasManifest() bool {
	manifestPath := fmt.Sprintf("%s/manifest.yaml", t.FullPath)
	_, err := os.Stat(manifestPath)
	return !os.IsNotExist(err)
}

func (t *UnigetTool) HasDockerfileTemplate() bool {
	dockerfilePath := fmt.Sprintf("%s/Dockerfile.template", t.FullPath)
	_, err := os.Stat(dockerfilePath)
	return !os.IsNotExist(err)
}

func (t *UnigetTool) LoadManifest() error {
	manifestPath := fmt.Sprintf("%s/manifest.yaml", t.FullPath)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("error reading manifest file: %w", err)
	}

	err = yaml.Unmarshal(data, &t.Tool)
	if err != nil {
		return fmt.Errorf("error unmarshalling manifest data: %w", err)
	}

	return nil
}

func (u *UnigetTools) FindTools() {
	entries, err := os.ReadDir(u.Directory)
	if err != nil {
		fmt.Printf("Failed to read directory: %v\n", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			toolName := entry.Name()

			t := NetUnigetTool(u.Directory, toolName)
			if !t.HasManifest() {
				fmt.Printf("Skipping %s: no manifest found\n", t.FullPath)
				continue
			}
			if !t.HasDockerfileTemplate() {
				fmt.Printf("Skipping %s: no Dockerfile.template found\n", t.FullPath)
				continue
			}

			err := t.LoadManifest()
			if err != nil {
				fmt.Printf("Error loading manifest for %s: %v\n", t.Subdirectory, err)
				continue
			}
			u.Tools[toolName] = *t
		}
	}
}

func (u *UnigetTools) Exists(toolName string) bool {
	_, exists := u.Tools[toolName]
	return exists
}
