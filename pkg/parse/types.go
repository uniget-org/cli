package parse

import (
	"bytes"
	"fmt"

	"github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/containers"
	"gitlab.com/uniget-org/cli/pkg/logging"
	"gitlab.com/uniget-org/cli/pkg/tool"
)

type ImageRefs struct {
	Refs       []ref.Ref
	BumpedRefs []ref.Ref
}

func (r *ImageRefs) Add(ref ref.Ref) {
	r.Refs = append(r.Refs, ref)
}

func (imageRefs *ImageRefs) Bump(tools *tool.Tools) error {
	imageRefs.BumpedRefs = make([]ref.Ref, len(imageRefs.Refs))

	for index, reference := range imageRefs.Refs {
		logging.Debugf("Bumping image reference: %s", reference)

		if reference.Registry == "ghcr.io" && reference.Repository[0:17] == "uniget-org/tools/" {
			toolName := reference.Repository[17:]
			tool, err := tools.GetByName(toolName)
			if err != nil {
				return fmt.Errorf("tool %s not found in metadata: %s", toolName, err)
			}

			reference.Tag = tool.Version
			reference.Digest = ""
			reference.Reference = fmt.Sprintf("%s/%s:%s", reference.Registry, reference.Repository, reference.Tag)
			reference.Digest, err = containers.FindNewDigest(reference)
			if err != nil {
				return fmt.Errorf("failed to find new digest for %s: %w", reference, err)
			}

			refReplacement := fmt.Sprintf("%s/%s:%s@%s", reference.Registry, reference.Repository, reference.Tag, reference.Digest)

			imageRefs.BumpedRefs[index], err = ref.New(refReplacement)
			if err != nil {
				return fmt.Errorf("failed to create ref from %q: %v", refReplacement, err)
			}

		} else {
			imageRefs.BumpedRefs[index] = reference
		}
	}

	return nil
}

func (imageRefs *ImageRefs) Replace(file []byte) ([]byte, error) {
	for i, ref := range imageRefs.Refs {
		file = bytes.ReplaceAll(file, []byte(ref.Reference), []byte(imageRefs.BumpedRefs[i].Reference))
	}

	return file, nil
}
