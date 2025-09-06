package parse

import "github.com/compose-spec/compose-go/v2/types"

func ExtractImageReferencesFromComposeFile(project *types.Project) (ImageRefs, error) {
	for _, service := range project.Services {
		if len(service.Image) > 0 {
			//
		}
	}

	return ImageRefs{}, nil
}
