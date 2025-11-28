package parse

import (
	"context"
	"fmt"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

func LoadComposeFile(composeFile string) (*types.Project, error) {
	options, err := cli.NewProjectOptions(
		[]string{composeFile},
		cli.WithOsEnv,
		cli.WithDotEnv,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose project options: %w", err)
	}

	project, err := options.LoadProject(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load compose project: %w", err)
	}

	return project, nil
}

func ExtractImageReferencesFromComposeFile(project *types.Project) (ImageRefs, error) {
	var imageRefs ImageRefs
	for _, service := range project.Services {
		if len(service.Image) > 0 {
			imageRef, err := ref.New(service.Image)
			if err != nil {
				logging.Debugf("Failed to create image reference from %s: %v", service.Image, err)
				continue
			}
			imageRefs.Add(imageRef)
		}
	}

	return imageRefs, nil
}

func BumpComposeFile(composeFile string, tools *tool.Tools) error {
	project, err := LoadComposeFile(composeFile)
	if err != nil {
		return fmt.Errorf("failed to load compose file: %w", err)
	}

	dockerfileNames := map[string]struct{}{}
	for _, service := range project.Services {
		if service.Build != nil {
			dockerfileName := service.Build.Dockerfile
			if service.Build.Dockerfile[0:1] != "/" {
				dockerfileName = service.Build.Context + "/" + service.Build.Dockerfile
			}
			dockerfileNames[dockerfileName] = struct{}{}
		}
	}
	for dockerfileName := range dockerfileNames {
		err := BumpDockerfile(dockerfileName, tools)
		if err != nil {
			return fmt.Errorf("failed to bump dockerfile %s: %w", dockerfileName, err)
		}
	}

	imageRefs, err := ExtractImageReferencesFromComposeFile(project)
	if err != nil {
		return fmt.Errorf("failed to extract image references: %w", err)
	}
	if len(imageRefs.Refs) == 0 {
		logging.Warning.Printfln("No image references found in compose file %s", composeFile)
		return nil
	}

	err = ReplaceInFile(composeFile, &imageRefs, tools)
	if err != nil {
		return fmt.Errorf("failed to replace image references in file: %w", err)
	}

	return nil
}
