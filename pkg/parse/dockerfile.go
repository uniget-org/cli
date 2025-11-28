package parse

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

func ExtractImageReferencesFromDockerfile(reader io.Reader) (ImageRefs, error) {
	result, err := parser.Parse(reader)
	if err != nil {
		return ImageRefs{}, fmt.Errorf("failed to parse Dockerfile: %w", err)
	}

	var imageRefs ImageRefs
	for _, child := range result.AST.Children {

		if strings.ToUpper(child.Value) == "FROM" && child.Next != nil {
			image := child.Next.Value

			imageRef, err := ref.New(image)
			if err != nil {
				logging.Debugf("Failed to create image reference from %s: %v", image, err)
				continue
			}
			imageRefs.Add(imageRef)
		}

		if strings.ToUpper(child.Value) == "COPY" || strings.ToUpper(child.Value) == "ADD" {
			for _, flag := range child.Flags {
				if flag[0:7] == "--from=" {
					image := flag[7:]

					fromRef, err := ref.New(image)
					if err != nil {
						logging.Debugf("Failed to create image reference from %s: %v", image, err)
						continue
					}

					imageRefs.Add(fromRef)
				}
			}
		}
	}

	return imageRefs, nil
}

func BumpDockerfile(dockerfile string, tools *tool.Tools) error {
	file, err := myos.SlurpFile(dockerfile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	reader := bytes.NewReader(file)
	imageRefs, err := ExtractImageReferencesFromDockerfile(reader)
	if err != nil {
		return fmt.Errorf("failed to extract image references: %w", err)
	}
	if len(imageRefs.Refs) == 0 {
		logging.Warning.Printfln("No image references found in Dockerfile %s", dockerfile)
		return nil
	}

	err = ReplaceInFile(dockerfile, &imageRefs, tools)
	if err != nil {
		return fmt.Errorf("failed to replace image references in file: %w", err)
	}

	return nil
}
