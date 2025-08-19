package parse

import (
	"fmt"
	"io"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/regclient/regclient/types/ref"
)

type ImageRefs struct {
	Refs []ref.Ref
}

func (r *ImageRefs) Add(ref ref.Ref) {
	r.Refs = append(r.Refs, ref)
}

func ExtractImageReferences(reader io.Reader) (ImageRefs, error) {
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
				return ImageRefs{}, fmt.Errorf("failed to create image reference: %w", err)
			}
			imageRefs.Add(imageRef)
		}

		if strings.ToUpper(child.Value) == "COPY" || strings.ToUpper(child.Value) == "ADD" {
			for _, flag := range child.Flags {
				if flag[0:7] == "--from=" {
					image := flag[7:]

					fromRef, err := ref.New(image)
					if err != nil {
						return ImageRefs{}, fmt.Errorf("failed to create image reference: %w", err)
					}

					imageRefs.Add(fromRef)
				}
			}
		}
	}

	return imageRefs, nil
}
